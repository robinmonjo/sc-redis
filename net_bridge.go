package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/docker/libcontainer/netlink"
)

const (
	bridgeName    string = "scredis0"
	bridgeNetwork string = "10.0.5.0/8" //must match with waht inside container_json.go
)

func setupNetBridge() error {
	// Enable IPv4 forwarding
	if err := ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte{'1', '\n'}, 0644); err != nil {
		return err
	}

	if err := netlink.CreateBridge(bridgeName, true); err != nil {
		// the bridge may already exist, therefore we can ignore an "exists" error
		if !os.IsExist(err) {
			return err
		}
	}

	iface, err := net.InterfaceByName(bridgeName)
	if err != nil {
		return err
	}

	ipAddr, ipNet, err := net.ParseCIDR(bridgeNetwork)
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

func bridgeInfo() string {
	return "bridge " + bridgeName + " up " + bridgeNetwork + ", (delete: `ip link delete " + bridgeName + " type bridge`)"
}
