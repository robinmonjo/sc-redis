package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/pkg/archive"
)

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
