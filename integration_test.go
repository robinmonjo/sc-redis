package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"
)

type binary struct {
	name   string
	addr   string
	ps     *os.Process
	stdout []byte
	stderr []byte
}

func newBinary(addr string) *binary {
	return &binary{name: "sc-redis", addr: addr}
}

func (sr *binary) start(args ...string) error {
	args = append([]string{"-w", os.TempDir()}, args...) //run test in /tmp
	cmd := exec.Command(sr.name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	sr.ps = cmd.Process
	sr.stdout, _ = ioutil.ReadAll(stdout)
	sr.stderr, _ = ioutil.ReadAll(stderr)
	return cmd.Wait()
}

func (sr *binary) waitUntilRunning() error {
	for i := 0; i < 30; i++ {
		c, err := net.DialTimeout("tcp", sr.addr, 3*time.Second)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		defer c.Close()
		return nil
	}
	return fmt.Errorf("unable to connect to redis")
}

func (sr *binary) stop() error {
	return sr.ps.Signal(syscall.SIGTERM)
}

func (sr *binary) printOutput() {
	fmt.Printf("Stdout: %s\nStderr: %s\n", string(sr.stdout), string(sr.stderr))
}

func Test_start(t *testing.T) {
	fmt.Printf("basic usage ... ")
	launch(t, newBinary("127.0.0.1:6379"))
	fmt.Println("done")
}

func Test_bridge(t *testing.T) {
	fmt.Printf("with net bridge ... ")
	launch(t, newBinary("10.0.5.22:6379"), "-i", "10.0.5.22")
	fmt.Println("done")
}

func Test_config(t *testing.T) {
	fmt.Printf("with net bridge and config ... ")
	launch(t, newBinary("10.0.5.66:6381"), "-i", "10.0.5.66", "-c", "port 6381")
	fmt.Println("done")
}

func Test_multi(t *testing.T) {
	fmt.Println("spawning 10 instances ...")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		port := fmt.Sprintf("909%d", i)
		go func(port string) {
			defer wg.Done()
			launch(t, newBinary("127.0.0.1:"+port), "-c", "port "+port)
			fmt.Printf("x ")
		}(port)
	}
	wg.Wait()
	fmt.Println("done")
}

func Test_multiBridge(t *testing.T) {
	fmt.Println("spawning 10 instances on the bridge ...")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		ip := fmt.Sprintf("10.0.5.%d", 200+i)
		go func(ip string) {
			defer wg.Done()
			launch(t, newBinary(ip+":6379"), "-i", ip)
			fmt.Printf("x ")
		}(ip)
	}
	wg.Wait()
	fmt.Println("done")
}

func launch(t *testing.T, b *binary, args ...string) {
	var (
		errStart   error
		errConnect error
		errStop    error
	)
	stopped := make(chan bool)
	go func() {
		errStart = b.start(args...)
		stopped <- true
	}()
	errConnect = b.waitUntilRunning()
	errStop = b.stop()

	<-stopped
	if errStart != nil || errConnect != nil || errStop != nil {
		fmt.Println("--------------------------------------------------------------------")
		fmt.Printf("\nStart error: %v\nConnect error: %v\nStop error: %v\n", errStart, errConnect, errStop)
		b.printOutput()
		fmt.Println("--------------------------------------------------------------------")
		t.FailNow()
	}
}
