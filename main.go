package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"text/template"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/namespaces"
	"github.com/fatih/color"
)

const (
	rootfsPath       = "./rootfs"
	redisRootfsAsset = "redis_rootfs.tar"
)

func init() {
	log.SetFlags(0) //no date time
}

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		//stage 2 execute inside container
		color.Set(color.FgGreen, color.Bold)
		log.SetPrefix("[Stage 2] ")
		log.Println("pid", os.Getpid(), "(inside container)") //will be pid one inside container

		_init()
		return
	}

	//stage 0 extracting rootfs
	color.Set(color.FgYellow, color.Bold)
	log.SetPrefix("[Stage 0] ")
	log.Println("pid", os.Getpid())
	log.Println("exporting redis container rootfs")
	if err := exportRootfs(); err != nil {
		log.Fatal(err)
	}

	ipLastInt, err := freeIpAddrLastInt()
	if err != nil {
		log.Fatal(err)
	}
	ipAddr := "10.0.5." + strconv.Itoa(ipLastInt) + "/8"
	if err := writeContainerJSON(ipAddr); err != nil {
		releaseIpAddr(ipLastInt)
		log.Fatal(err)
	}
	color.Unset()

	//stage 1 starting container
	color.Set(color.FgBlue, color.Bold)
	log.SetPrefix("[Stage 1] ")
	if err := setupNetBridge(); err != nil {
		releaseIpAddr(ipLastInt)
		log.Fatal(err)
	}
	log.Println(bridgeInfo())
	log.Println("container IP address:", ipAddr)
	log.Println("starting container")
	container, err := loadConfig(rootfsPath)
	if err != nil {
		releaseIpAddr(ipLastInt)
		log.Fatal(err)
	}

	exitCode, err := startContainer(container, rootfsPath, []string{"redis-server"})

	if err != nil {
		releaseIpAddr(ipLastInt)
		log.Fatalf("failed to exec: %s", err)
	}
	color.Unset()

	releaseIpAddr(ipLastInt)
	os.Exit(exitCode)
}

func exportRootfs() error {
	//export the tar
	tar, err := Asset(redisRootfsAsset)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(tar)

	return archive.Untar(buf, rootfsPath, nil)
}

func writeContainerJSON(ipAddr string) error {
	//write the container.json
	f, err := os.Create(path.Join(rootfsPath, "container.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	type Config struct {
		IpAddr string
	}

	t := template.New("container.json")
	t, err = t.Parse(containerJson)
	if err != nil {
		return err
	}
	t.Execute(f, Config{IpAddr: ipAddr})
	return nil
}

// startContainer starts the container. Returns the exit status or -1 and an
// error.
//
// Signals sent to the current process will be forwarded to container.
func startContainer(container *libcontainer.Config, dataPath string, args []string) (int, error) {
	var (
		cmd     *exec.Cmd
		sigc    = make(chan os.Signal, 10)
		console = ""
	)

	signal.Notify(sigc)

	createCommand := func(container *libcontainer.Config, console, dataPath, init string, pipe *os.File, args []string) *exec.Cmd {
		cmd = namespaces.DefaultCreateCommand(container, console, dataPath, init, pipe, args)
		return cmd
	}

	startCallback := func() {
		go func() {
			for sig := range sigc {
				cmd.Process.Signal(sig)
			}
		}()
	}

	return namespaces.Exec(container, os.Stdin, os.Stdout, os.Stderr, console, dataPath, args, createCommand, startCallback)
}

//container pid 1 code
func _init() {
	err := os.Chdir(rootfsPath)
	if err != nil {
		log.Fatal(err)
	}

	var (
		console   = os.Getenv("console")
		rawPipeFd = os.Getenv("pipe")
	)
	runtime.LockOSThread()

	rootfs, err := os.Getwd()
	log.Println("container is in ", rootfs)
	if err != nil {
		log.Fatal(err)
	}

	container, err := loadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	pipeFd, err := strconv.Atoi(rawPipeFd)
	if err != nil {
		log.Fatal(err)
	}

	pipe := os.NewFile(uintptr(pipeFd), "pipe")
	args := findUserArgs(os.Args)
	log.Println("executing", args)
	color.Unset()
	if err := namespaces.Init(container, rootfs, console, pipe, args); err != nil {
		log.Fatalf("unable to initialize container: %s", err)
	}
}
