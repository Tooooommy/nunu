//go:build !plan9 && !windows
// +build !plan9,!windows

package run

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func killProcess(cmd *exec.Cmd) error {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

func start(dir string, programArgs []string) *exec.Cmd {
	cmd := exec.Command("go", append([]string{"run", dir}, programArgs...)...)
	// Set a new process group to kill all child processes when the program exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("\033[33;1mcmd run failed\u001B[0m")
	}
	time.Sleep(time.Second)
	fmt.Printf("\033[32;1mrunning...\033[0m\n")
	return cmd
}
