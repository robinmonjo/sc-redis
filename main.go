package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"

	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/namespaces"
)

const rootfs string = "."

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		_init()
		return
	}

	container, err := loadConfig(rootfs)
	if err != nil {
		log.Fatal(err)
	}

	exitCode, err := startContainer(container, rootfs, []string{"/usr/local/bin/redis-server"})

	if err != nil {
		log.Fatalf("failed to exec: %s", err)
	}

	os.Exit(exitCode)
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

//called by the contained process.
func _init() {
	var (
		console   = os.Getenv("console")
		rawPipeFd = os.Getenv("pipe")
	)
	runtime.LockOSThread()

	rootfs, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	container, err := loadConfig(rootfs)
	if err != nil {
		log.Fatal(err)
	}

	pipeFd, err := strconv.Atoi(rawPipeFd)
	if err != nil {
		log.Fatal(err)
	}

	pipe := os.NewFile(uintptr(pipeFd), "pipe")

	if err := namespaces.Init(container, rootfs, console, pipe, findUserArgs(os.Args)); err != nil {
		log.Fatalf("unable to initialize for container: %s", err)
	}
}
