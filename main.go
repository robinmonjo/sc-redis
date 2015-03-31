package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/utils"
)

const (
	//versions
	version             = "1.1.1"
	libcontainerVersion = "b6cf7a6c8520fd21e75f8b3becec6dc355d844b0"
	redisVersion        = "2.8.19"

	//bridge
	vethBridge  = "scredis0"
	vethNetwork = "172.18.1.1/16"
	vethGateway = "172.18.1.1"
)

func init() {
	log.SetFlags(0) //no date time
}

func main() {
	app := cli.NewApp()
	app.Name = "sc-redis"
	app.Version = fmt.Sprintf("v%s (redis v%s, libcontainer %s)", version, redisVersion, libcontainerVersion)
	app.Author = "Robin Monjo"
	app.Email = "robinmonjo@gmail.com"
	app.Usage = "self contained redis-server"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "config, c", Usage: "redis configuration, e.g: \"requirepass foobar, port 9999, ...\""},
		cli.StringFlag{Name: "ip, i", Usage: "use the net namespace with the given ip address, format: 172.18.xxx.xxx"},
		cli.StringFlag{Name: "working_dir, w", Value: ".", Usage: "working directory where container are created"},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "init",
			Usage:  "container init, should never be invoked manually",
			Action: initAction,
		},
	}
	app.Action = func(c *cli.Context) {
		exit, err := start(c)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(exit)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initAction(c *cli.Context) {
	log.SetPrefix("[container] ")

	log.Println("pid", os.Getpid()) //will be pid one inside container
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()

	factory, err := libcontainer.New("")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("starting redis")
	if err := factory.StartInitialization(3); err != nil {
		log.Fatal(err)
	}
	panic("This line should never been executed")
}

func start(c *cli.Context) (int, error) {
	log.SetPrefix("[host] ")

	uid, err := utils.GenerateRandomName("sc_redis_", 7)
	if err != nil {
		return 1, err
	}

	log.Println("pid", os.Getpid())
	log.Println("container uid:", uid)
	log.Println("exporting container rootfs")

	rootfs := path.Join(c.GlobalString("working_dir"), uid)
	rootfs, _ = filepath.Abs(rootfs)

	defer os.RemoveAll(rootfs)
	if err := exportRootfs(rootfs); err != nil {
		return 1, err
	}

	log.Println("writing redis configuration")
	if err := writeRawRedisConf(path.Join(rootfs, "etc"), c.GlobalString("config")); err != nil {
		return 1, err
	}

	ipAddr := c.GlobalString("ip")
	if ipAddr != "" {
		if err := setupNetBridge(); err != nil {
			return 1, err
		}
		log.Println("bridge " + vethBridge + " up " + vethNetwork)
		if err := validateIPAddr(ipAddr); err != nil {
			return 1, err
		}
		log.Println("container IP address:", ipAddr)
		ipAddr = ipAddr + "/8"
	}

	factory, err := libcontainer.New(rootfs)
	if err != nil {
		return 1, err
	}

	container, err := factory.Create(uid, loadConfig(uid, rootfs, ipAddr))
	if err != nil {
		return 1, err
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
		return 1, err
	}

	// wait for the process to finish.
	status, err := process.Wait()
	if err != nil {
		return 1, err
	}

	// destroy the container.
	log.Println("Cleaning up")
	container.Destroy()

	return utils.ExitStatus(status.Sys().(syscall.WaitStatus)), nil
}

func handleSignals(container *libcontainer.Process) {
	sigc := make(chan os.Signal, 10)
	signal.Notify(sigc)
	for sig := range sigc {
		container.Signal(sig)
	}
}
