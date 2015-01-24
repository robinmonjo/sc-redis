package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/libcontainer"
)

const ipPoolFile = "/etc/scredis_ips.json"

func loadConfig(rootfs string) (*libcontainer.Config, error) {
	f, err := os.Open(filepath.Join(rootfs, "container.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var container *libcontainer.Config
	if err := json.NewDecoder(f).Decode(&container); err != nil {
		return nil, err
	}

	return container, nil
}

func findUserArgs(args []string) []string {
	i := 0
	for _, a := range args {
		i++

		if a == "--" {
			break
		}
	}

	return args[i:]
}

func loadUsedIpAddrLastInts() ([]int, error) {
	if _, err := os.Stat(ipPoolFile); os.IsNotExist(err) {
		if err := ioutil.WriteFile(ipPoolFile, []byte("[]\n"), 0666); err != nil {
			return nil, err
		}
	}

	f, err := os.Open(ipPoolFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ipLastInts []int
	if err := json.NewDecoder(f).Decode(&ipLastInts); err != nil {
		return nil, err
	}
	return ipLastInts, nil
}

func freeIpAddrLastInt() (int, error) {
	ipLastInts, err := loadUsedIpAddrLastInts()
	if err != nil {
		return -1, err
	}

	bitVector := make([]bool, 252) //range is 2 .. 254
	for _, i := range ipLastInts {
		bitVector[i-2] = true
	}

	freeIpLastInt := -1
	for i, taken := range bitVector {
		if !taken {
			freeIpLastInt = i + 2
			break
		}
	}
	if freeIpLastInt == -1 {
		return -1, fmt.Errorf("no more ip addr available")
	}

	ipLastInts = append(ipLastInts, freeIpLastInt)
	b, _ := json.Marshal(ipLastInts)
	if err := ioutil.WriteFile(ipPoolFile, b, 0666); err != nil {
		return -1, err
	}

	return freeIpLastInt, nil
}

func releaseIpAddr(ipLastInt int) error {
	ipLastInts, err := loadUsedIpAddrLastInts()
	if err != nil {
		return err
	}
	idx := -1
	for i, v := range ipLastInts {
		if v == ipLastInt {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}
	ipLastInts = append(ipLastInts[:idx], ipLastInts[idx+1:]...)
	b, _ := json.Marshal(ipLastInts)
	return ioutil.WriteFile(ipPoolFile, b, 0666)
}
