//go:build windows

package main

import (
	"os"
	"os/exec"
)

// TriggerRestart spawns the new binary and terminates the current process
func TriggerRestart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exePath)
	err = cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
