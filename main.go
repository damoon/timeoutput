package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {

	if len(os.Args) < 4 {
		log.Fatalf("usage: %s [global timeout] [output timeout] [command]", os.Args[0])
	}

	globalTimeout, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatalf("failed to get timeout: %s", err)
	}
	outputTimeout, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		log.Fatalf("failed to get timeoutput: %s", err)
	}
	args := os.Args[3:]

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = outW
	cmd.Stderr = errW

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		for s := range gracefulStop {
			cmd.Process.Signal(s)
		}
	}()

	globalDuration := time.Duration(globalTimeout) * time.Second
	outputDuration := time.Duration(outputTimeout) * time.Second

	globalTimer := time.NewTimer(globalDuration)
	outputTimer := time.NewTimer(outputDuration)

	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start process: %s", err)
	}

	go transfer(outR, os.Stdout, outputTimer, outputDuration)
	go transfer(errR, os.Stderr, outputTimer, outputDuration)

	go func() {
		select {
		case <-globalTimer.C:
			shutdown(globalTimer, outputTimer, cmd)
		case <-outputTimer.C:
			shutdown(globalTimer, outputTimer, cmd)
		}
	}()

	err = cmd.Wait()
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

func shutdown(gt, ot *time.Timer, c *exec.Cmd) {
	gt.Stop()
	ot.Stop()
	err := c.Process.Signal(syscall.SIGHUP)
	if err != nil {
		log.Printf("failed to signal to process: %s", err)
	}
	go func() {
		time.Sleep(time.Second)
		err = c.Process.Kill()
		if err != nil {
			log.Printf("failed to kill process: %s", err)
		}
	}()
}

func transfer(r io.Reader, w io.WriteCloser, t *time.Timer, d time.Duration) {
	for {
		b := make([]byte, 1024)
		n, err := r.Read(b)
		if err == io.EOF {
			w.Close()
			break
		}
		if err != nil {
			log.Fatalf("failed to read from process: %s", err)
		}
		if !t.Stop() {
			<-t.C
		}
		t.Reset(d)
		w.Write(b[:n])
	}
}
