package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/libcontainer"
)

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

func validateIPAddr(ip string) error {
	formatErr := fmt.Errorf("invalid ip address %s. Expecting 10.0.5.XXX", ip)

	comps := strings.Split(ip, ".")
	if len(comps) != 4 {
		return formatErr
	}
	if comps[0] != "10" || comps[1] != "0" || comps[2] != "5" {
		return formatErr
	}
	ipID, err := strconv.Atoi(comps[3])
	if err != nil {
		return err
	}
	if ipID < 2 || ipID > 254 {
		return fmt.Errorf("%s out of ip range (2..254)", comps[3])
	}
	return nil
}

func exportRootfs(basePath string) error {
	tar, err := Asset("redis_rootfs.tar")
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(tar)

	return archive.Untar(buf, basePath, nil)
}
