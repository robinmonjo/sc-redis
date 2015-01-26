package main

import (
	"bytes"
	"os"
	"path"
	"text/template"
)

const containerJson = `
{
    "capabilities": [
        "CHOWN",
        "DAC_OVERRIDE",
        "FOWNER",
        "MKNOD",
        "NET_RAW",
        "SETGID",
        "SETUID",
        "SETFCAP",
        "SETPCAP",
        "NET_BIND_SERVICE",
        "SYS_CHROOT",
        "KILL"
    ],
    "cgroups": {
        "allowed_devices": [
            {
                "cgroup_permissions": "m",
                "major_number": -1,
                "minor_number": -1,
                "type": 99
            },
            {
                "cgroup_permissions": "m",
                "major_number": -1,
                "minor_number": -1,
                "type": 98
            },
            {
                "cgroup_permissions": "rwm",
                "major_number": 5,
                "minor_number": 1,
                "path": "/dev/console",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "major_number": 4,
                "path": "/dev/tty0",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "major_number": 4,
                "minor_number": 1,
                "path": "/dev/tty1",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "major_number": 136,
                "minor_number": -1,
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "major_number": 5,
                "minor_number": 2,
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "major_number": 10,
                "minor_number": 200,
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 3,
                "path": "/dev/null",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 5,
                "path": "/dev/zero",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 7,
                "path": "/dev/full",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 5,
                "path": "/dev/tty",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 9,
                "path": "/dev/urandom",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 8,
                "path": "/dev/random",
                "type": 99
            }
        ],
        "name": "redis",
        "parent": "sc-redis"
    },
    "restrict_sys": true,
    "mount_config": {
        "device_nodes": [
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 3,
                "path": "/dev/null",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 5,
                "path": "/dev/zero",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 7,
                "path": "/dev/full",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 5,
                "path": "/dev/tty",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 9,
                "path": "/dev/urandom",
                "type": 99
            },
            {
                "cgroup_permissions": "rwm",
                "file_mode": 438,
                "major_number": 1,
                "minor_number": 8,
                "path": "/dev/random",
                "type": 99
            }
        ],
        "mounts": [
            {
                "type": "tmpfs",
                "destination": "/tmp"
            }
        ]
    },
    "environment": [
        "HOME=/",
        "PATH=/usr/local/bin",
        "HOSTNAME=redis",
        "TERM=xterm"
    ],
    "hostname": "redis",
    "namespaces": {
        "NEWIPC": true,
        "NEWNS": true,
        "NEWPID": true,
        {{if .NetNamespace}} {{ .NetNamespace }} {{end}}
        "NEWUTS": true
    },
    {{if .NetIfaces}} {{ .NetIfaces }} {{end}}
    "tty": false,
    "user": "root"
}
`

const networkIfaces = `
"networks": [
    {
        "address": "127.0.0.1/0",
        "gateway": "localhost",
        "mtu": 1500,
        "type": "loopback"
    },
    {
        "address": "{{.IpAddr}}",
        "bridge": "scredis0",
        "veth_prefix": "veth",
        "gateway": "10.0.5.1",
        "mtu": 1500,
        "type": "veth"
    }
],
`

const networkNamespace = `
  "NEWNET": true,
`

//if ipAddr == "", will use host network otherwise, will setup the net namespace
func writeContainerJSON(rootfs, ipAddr string) error {
	f, err := os.Create(path.Join(rootfs, "container.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	m := make(map[string]string)
	m["IpAddr"] = ipAddr

	if ipAddr != "" {
		tIfaces := template.New("networks")
		tIfaces, err = tIfaces.Parse(networkIfaces)
		if err != nil {
			return err
		}
		var netIfaces bytes.Buffer
		if err := tIfaces.Execute(&netIfaces, map[string]string{"IpAddr": ipAddr}); err != nil {
			return err
		}
		m["NetIfaces"] = netIfaces.String()
		m["NetNamespace"] = networkNamespace
	}

	t := template.New("container.json")
	t, err = t.Parse(containerJson)
	if err != nil {
		return err
	}
	//var tmp bytes.Buffer
	return t.Execute(f, m)
	//log.Println(tmp.String())
	//return nil
}
