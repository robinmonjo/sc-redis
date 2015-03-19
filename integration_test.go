package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/fzzy/radix/redis"
)

type binary struct {
	name string
	addr string
	ps   *os.Process
}

func newBinary(addr string) *binary {
	return &binary{name: "sc-redis", addr: addr}
}

func (sr *binary) start(args ...string) error {
	args = append([]string{"-w", os.TempDir()}, args...) //run test is /tmp
	cmd := exec.Command(sr.name, args...)
	cmd.Start()
	sr.ps = cmd.Process
	return cmd.Wait()
}

func (sr *binary) waitUntilRunning() error {
	for i := 0; i < 5; i++ {
		c, err := redis.DialTimeout("tcp", sr.addr, 3*time.Second)
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

func launch(t *testing.T, b *binary, args ...string) {
	var err error
	stopped := make(chan bool)
	go func() {
		err = b.start(args...)
		stopped <- true
	}()
	if err := b.waitUntilRunning(); err != nil {
		t.Fatal(err)
	}
	if err := b.stop(); err != nil {
		t.Fatal(err)
	}
	<-stopped
	if err != nil {
		t.Fatal(err)
	}
}
