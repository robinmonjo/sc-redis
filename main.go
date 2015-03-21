package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/utils"
)

const (
	//versions
	version             = "0.2"
	libcontainerVersion = "1.4.0"
	redisVersion        = "2.8.19"

	//bridge
	vethBridge  = "scredis0"
	vethNetwork = "10.0.5.0/8"
	vethGateway = "10.0.5.1"
)

var (
	showVersion *bool   = flag.Bool("v", false, "show version")
	redisConfig *string = flag.String("c", "", "specify specific redis configuration")
	bridgedIP   *string = flag.String("i", "", "use the net namespace with this ip (format: 10.0.5.XXX)")
	workingDir  *string = flag.String("w", ".", "working directory")
	rootfsName  *string = flag.String("n", "", "name of the rootfs dir")
)

func init() {
	log.SetFlags(0) //no date time
}

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		//stage 1 execute inside container
		log.SetPrefix("[Stage 1] ")
		log.Println("pid", os.Getpid(), "(inside container)") //will be pid one inside container

		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()

		factory, err := libcontainer.New("")
		if err != nil {
			log.Fatal(err)
		}
		if err := factory.StartInitialization(3); err != nil {
			log.Fatal(err)
		}
		panic("This line should never been executed")
	}

	flag.Parse()

	if *showVersion {
		log.Printf("sc-redis v%s (redis v%s, libcontainer v%s)", version, redisVersion, libcontainerVersion)
		os.Exit(0)
	}

	name := *rootfsName
	if name == "" {
		name = "scredis_" + time.Now().Format("20060102150405")
	}
	rootfs := path.Join(*workingDir, name)
	rootfs, _ = filepath.Abs(rootfs)

	//stage 0 extracting rootfs
	log.SetPrefix("[Stage 0] ")
	log.Println("pid", os.Getpid())
	log.Println("exporting redis container rootfs")
	if err := exportRootfs(rootfs); err != nil {
		log.Fatal(err)
	}

	if err := writeRawRedisConf(path.Join(rootfs, "etc"), *redisConfig); err != nil {
		log.Fatal(err)
	}
	log.Println("writing redis configuration")

	ipAddr := *bridgedIP
	if ipAddr != "" {
		if err := setupNetBridge(); err != nil {
			log.Fatal(err)
		}
		log.Println("bridge " + vethBridge + " up " + vethNetwork)
		if err := validateIPAddr(ipAddr); err != nil {
			log.Fatal(err)
		}
		log.Println("container IP address:", ipAddr)
		ipAddr = ipAddr + "/8"
	}

	factory, err := libcontainer.New(rootfs)
	if err != nil {
		log.Fatal(err)
	}

	contId, err := utils.GenerateRandomName("sc_redis_", 7)
	if err != nil {
		log.Fatal(err)
	}

	container, err := factory.Create(contId, loadConfig(rootfs, ipAddr))
	if err != nil {
		log.Fatal(err)
	}
	process := &libcontainer.Process{
		Args:   []string{"redis-server", "/etc/redis.conf"},
		Env:    []string{"PATH=/usr/local/bin"},
		User:   "root",
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	go handleSignals(process)

	if err := container.Start(process); err != nil {
		log.Fatal(err)
	}

	// wait for the process to finish.
	status, err := process.Wait()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(status)

	// destroy the container.
	log.Println("Cleaning up")
	container.Destroy()
	if err := os.RemoveAll(rootfs); err != nil {
		log.Fatal(err)
	}

	os.Exit(utils.ExitStatus(status.Sys().(syscall.WaitStatus)))
}

func handleSignals(container *libcontainer.Process) {
	sigc := make(chan os.Signal, 10)
	signal.Notify(sigc)
	for sig := range sigc {
		container.Signal(sig)
	}
}
