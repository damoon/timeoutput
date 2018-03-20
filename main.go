package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	args := os.Args[1:]

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//	cmd.Process.Signal("HUP")

	//	cmd.Process.Kill()

	err := cmd.Run()
	//	log.Fatalln("sub process started")

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			os.Exit(waitStatus.ExitStatus())
		}
		fmt.Fprintf(os.Stderr, "execution failed: %v\n", err)
		os.Exit(127)
	}

	os.Exit(0)
}
