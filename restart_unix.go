//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// TriggerRestart spawns the new binary in a detached session and terminates the current process
func TriggerRestart() error {
	exePath := originalExePath
	if exePath == "" {
		var err error
		exePath, err = os.Executable()
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(exePath)
	// Detach the child process from the parent's process group/session
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err := cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
