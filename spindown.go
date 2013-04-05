package main

import (
	"os"
	"os/exec"
)

func spindown(dev string) error {
	cmd := exec.Command("hdparm", "-y", "--", dev)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
