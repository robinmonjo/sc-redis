package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
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

func loadInUseIPAddrIDs() ([]int, error) {
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

	var ipAddrIDs []int
	if err := json.NewDecoder(f).Decode(&ipAddrIDs); err != nil {
		return nil, err
	}
	return ipAddrIDs, nil
}

func availableIPAddrID() (int, error) {
	ipAddrIDs, err := loadInUseIPAddrIDs()
	if err != nil {
		return -1, err
	}

	bitVector := make([]bool, 252) //range is 2 .. 254
	for _, i := range ipAddrIDs {
		bitVector[i-2] = true
	}

	availableIPAddrID := -1
	for i, inUse := range bitVector {
		if !inUse {
			availableIPAddrID = i + 2
			break
		}
	}
	if availableIPAddrID == -1 {
		return -1, fmt.Errorf("no more ip addr available")
	}

	ipAddrIDs = append(ipAddrIDs, availableIPAddrID)
	b, _ := json.Marshal(ipAddrIDs)
	if err := ioutil.WriteFile(ipPoolFile, b, 0666); err != nil {
		return -1, err
	}

	return availableIPAddrID, nil
}

func releaseIpAddr(ipID int) error {
	ipAddrIDs, err := loadInUseIPAddrIDs()
	if err != nil {
		return err
	}
	idx := -1
	for i, v := range ipAddrIDs {
		if v == ipID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}
	ipAddrIDs = append(ipAddrIDs[:idx], ipAddrIDs[idx+1:]...)
	b, _ := json.Marshal(ipAddrIDs)
	return ioutil.WriteFile(ipPoolFile, b, 0666)
}

func exportRootfs(basePath string) error {
	//export the tar
	tar, err := Asset(redisRootfsAsset)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(tar)

	return archive.Untar(buf, basePath, nil)
}
