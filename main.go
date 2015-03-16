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
	"time"

	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/namespaces"
)

const (
	version             = "0.2"
	libcontainerVersion = "1.4.0"
	redisVersion        = "2.8.19"
	bridgeName          = "scredis0"
	bridgeNetwork       = "10.0.5.0/8" //must match with what inside container_json.go
)

var (
	showVersion *bool   = flag.Bool("v", false, "show version")
	redisConfig *string = flag.String("c", "", "specify specific redis configuration")
	bridgedIP   *string = flag.String("i", "", "use the net namespace with this ip (format: 10.0.5.XXX)")
)

func init() {
	log.SetFlags(0) //no date time
}

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		//stage 2 execute inside container
		log.SetPrefix("[Stage 2] ")
		log.Println("pid", os.Getpid(), "(inside container)") //will be pid one inside container

		initProcess()
		return
	}

	flag.Parse()

	if *showVersion {
		log.Printf("sc-redis v%s (redis v%s, libcontainer v%s)", version, redisVersion, libcontainerVersion)
		os.Exit(0)
	}

	rootfs := "scredis_" + time.Now().Format("20060102150405")
	startContFn, err := prepareContainer(rootfs, *bridgedIP)
	if err != nil {
		log.Fatal(err)
	}
	exitCode, err := startContFn()
	if err != nil {
		log.Fatal(err)
	}

	os.RemoveAll(rootfs)
	os.Exit(exitCode)
}

//Prepare container rootfs + return the function to start it.
//If ipAddr == "", will use host network, otherwise will setup net namelspace
func prepareContainer(rootfs, ipAddr string) (func() (int, error), error) {
	var (
		err           error
		containerConf *libcontainer.Config
	)

	defer func() {
		//if something goes wrong during preparation, cleanup
		if err != nil {
			os.RemoveAll(rootfs)
		}
	}()

	//stage 0 extracting rootfs
	log.SetPrefix("[Stage 0] ")
	log.Println("pid", os.Getpid())
	log.Println("exporting redis container rootfs")
	if err = exportRootfs(rootfs); err != nil {
		return nil, err
	}

	//stage 1 configuring container
	log.SetPrefix("[Stage 1] ")

	if ipAddr != "" {
		if err = setupNetBridge(); err != nil {
			return nil, err
		}
		log.Println("bridge " + bridgeName + " up " + bridgeNetwork)
		if err = validateIPAddr(ipAddr); err != nil {
			return nil, err
		}
		log.Println("container IP address:", ipAddr)
		ipAddr = ipAddr + "/8"
	}

	if err = writeContainerJSON(rootfs, ipAddr); err != nil {
		return nil, err
	}
	log.Println("container config exported")
	if err = writeRawRedisConf(path.Join(rootfs, "etc"), *redisConfig); err != nil {
		return nil, err
	}
	log.Println("redis config exported")

	containerConf, err = loadConfig(rootfs)
	if err != nil {
		return nil, err
	}
	fn := func() (int, error) {
		return startContainer(containerConf, rootfs, []string{"redis-server", "/etc/redis.conf"})
	}

	return fn, nil
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
	log.Println("container is in", rootfs)
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
	if err := namespaces.Init(container, rootfs, console, pipe, args); err != nil {
		log.Fatalf("unable to initialize container: %s", err)
	}
}
