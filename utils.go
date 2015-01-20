package main

import (
	"encoding/json"
	"os"
	"path/filepath"

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
