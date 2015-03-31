package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/libcontainer/netlink"
)

func setupNetBridge() error {
	// Enable IPv4 forwarding
	if err := ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte{'1', '\n'}, 0644); err != nil {
		return err
	}

	if err := netlink.CreateBridge(vethBridge, false); err != nil {
		// the bridge may already exist, therefore we can ignore an "exists" error
		if !os.IsExist(err) {
			return err
		}
	}

	iface, err := net.InterfaceByName(vethBridge)
	if err != nil {
		return err
	}

	ipAddr, ipNet, err := net.ParseCIDR(vethNetwork)
	if err != nil {
		return err
	}

	if netlink.NetworkLinkAddIp(iface, ipAddr, ipNet); err != nil {
		return fmt.Errorf("failed to add private network: %s", err)
	}
	if err := netlink.NetworkLinkUp(iface); err != nil {
		return fmt.Errorf("failed to start network bridge: %s", err)
	}

	return nil
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
