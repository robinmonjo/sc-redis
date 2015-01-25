package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"

	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/namespaces"
	"github.com/fatih/color"
)

const (
	redisRootfsAsset = "redis_rootfs.tar"
	version          = "0.1"
	dockerVersion    = "1.4.1"
)

var (
	showVersion *bool   = flag.Bool("v", false, "show version")
	redisConfig *string = flag.String("c", "", "specify specific redis configuration")
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

		initProcess()
		return
	}

	flag.Parse()

	if *showVersion {
		log.Printf("sc-redis v%s (docker v%s)", version, dockerVersion)
		os.Exit(0)
	}

	rootfs := "./sc-redis_rootfs"
	startContFn, ipID, err := prepareContainer(rootfs)
	exitCode, err := startContFn()
	if err != nil {
		color.Unset()
		log.Fatal(err)
	}
	releaseIpAddr(ipID)
	os.RemoveAll("./sc-redis_rootfs")
	os.Exit(exitCode)
}

//Prepare container rootfs + return the function to start it
func prepareContainer(rootfs string) (func() (int, error), int, error) {
	var (
		err           error
		ipID          int
		containerConf *libcontainer.Config
	)

	defer func() {
		//if something went wrong during preparation, cleanup
		if err != nil {
			os.RemoveAll(rootfs)
			releaseIpAddr(ipID)
			color.Unset()
		}
	}()

	//stage 0 extracting rootfs
	color.Set(color.FgYellow, color.Bold)
	log.SetPrefix("[Stage 0] ")
	log.Println("pid", os.Getpid())
	log.Println("exporting redis container rootfs")
	if err = exportRootfs(rootfs); err != nil {
		return nil, ipID, err
	}

	ipID, err = availableIPAddrID()
	if err != nil {
		return nil, ipID, err
	}
	ipAddr := "10.0.5." + strconv.Itoa(ipID) + "/8"
	if err = writeContainerJSON(rootfs, ipAddr); err != nil {
		return nil, ipID, err
	}
	if err = writeRawRedisConf(path.Join(rootfs, "etc"), *redisConfig); err != nil {
		return nil, ipID, err
	}

	color.Unset()

	//stage 1 starting container
	color.Set(color.FgBlue, color.Bold)
	log.SetPrefix("[Stage 1] ")
	if err = setupNetBridge(); err != nil {
		return nil, ipID, err
	}
	log.Println(bridgeInfo())
	log.Println("container IP address:", ipAddr)
	log.Println("starting container")
	containerConf, err = loadConfig(rootfs)
	if err != nil {
		return nil, ipID, err
	}

	fn := func() (int, error) {
		return startContainer(containerConf, rootfs, []string{"redis-server", "/etc/redis.conf"})
	}

	return fn, ipID, nil
}

// startContainer starts the container. Returns the exit status or -1 and an
// error.
//
// Signals sent to the current process will be forwarded to container.
func startContainer(container *libcontainer.Config, rootfs string, args []string) (int, error) {
	var (
		cmd     *exec.Cmd
		sigc    = make(chan os.Signal, 10)
		console = ""
	)

	signal.Notify(sigc)

	createCommand := func(container *libcontainer.Config, console, rootfs, init string, pipe *os.File, args []string) *exec.Cmd {
		cmd = namespaces.DefaultCreateCommand(container, console, rootfs, init, pipe, args)
		cmd.Env = append(cmd.Env, "rootfs="+rootfs)
		return cmd
	}

	startCallback := func() {
		go func() {
			for sig := range sigc {
				cmd.Process.Signal(sig)
			}
		}()
	}

	return namespaces.Exec(container, os.Stdin, os.Stdout, os.Stderr, console, rootfs, args, createCommand, startCallback)
}

//container pid 1 code
func initProcess() {
	var (
		console   = os.Getenv("console")
		rawPipeFd = os.Getenv("pipe")
		rootfs    = os.Getenv("rootfs")
	)

	err := os.Chdir(rootfs)
	if err != nil {
		log.Fatal(err)
	}

	runtime.LockOSThread()

	rootfs, err = os.Getwd()
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
