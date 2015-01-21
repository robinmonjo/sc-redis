package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/namespaces"
	"github.com/fatih/color"
)

const (
	rootfsPath       string = "./rootfs"
	redisRootfsAsset        = "redis_rootfs.tar"
)

func init() {
	log.SetFlags(0) //no date time
}

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		//stage 2 execute inside container
		color.Set(color.FgGreen, color.Bold)
		log.SetPrefix("[Stage 2] ")
		_init()
		return
	}

	//stage 0 extracting rootfs
	color.Set(color.FgYellow, color.Bold)
	log.SetPrefix("[Stage 0] ")
	log.Println("exporting redis container rootfs")
	exportRootfs()
	color.Unset()

	//stage 1 starting container
	color.Set(color.FgBlue, color.Bold)
	log.SetPrefix("[Stage 1] ")
	log.Println("starting container")
	container, err := loadConfig(rootfsPath)
	if err != nil {
		log.Fatal(err)
	}

	exitCode, err := startContainer(container, rootfsPath, []string{"redis-server"})

	if err != nil {
		log.Fatalf("failed to exec: %s", err)
	}
	color.Unset()

	os.Exit(exitCode)
}

func exportRootfs() {
	//export the tar
	tar, err := Asset(redisRootfsAsset)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(tar)

	err = archive.Untar(buf, rootfsPath, nil)
	if err != nil {
		log.Fatal(err)
	}

	//write the container.json
	err = ioutil.WriteFile(path.Join(rootfsPath, "container.json"), []byte(containerJson), 0644)
	if err != nil {
		log.Fatal(err)
	}
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
		log.Fatalf("unable to initialize for container: %s", err)
	}
}
