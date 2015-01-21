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
)

const (
	rootfsPath       string = "./rootfs"
	redisRootfsAsset        = "redis_rootfs.tar"
)

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		_init()
		return
	}

	exportRootfs()

	container, err := loadConfig(rootfsPath)
	if err != nil {
		log.Fatal(err)
	}

	exitCode, err := startContainer(container, rootfsPath, []string{"redis-server"})

	if err != nil {
		log.Fatalf("failed to exec: %s", err)
	}

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

//called by the contained process.
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

	if err := namespaces.Init(container, rootfs, console, pipe, findUserArgs(os.Args)); err != nil {
		log.Fatalf("unable to initialize for container: %s", err)
	}
}
