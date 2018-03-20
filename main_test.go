package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

var mainTests = []struct {
	args     []string
	stdin    string
	stdout   string
	stderr   string
	exitcode int
}{
	{[]string{"true"}, "", "", "", 0},
	{[]string{"false"}, "", "", "exit status 1\n", 1},
	{[]string{"echo", "-n", "Hello, world!"}, "", "Hello, world!", "", 0},
	{[]string{"wc", "-c"}, "Hello, world!", "13\n", "", 0},
}

func TestTrue(t *testing.T) {

	for _, tt := range mainTests {

		args := []string{"go", "run", "main.go"}
		args = append(args, tt.args...)

		cmd := exec.Command(args[0], args[1:]...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			t.Errorf("failed to prepare command input: %s", err)
		}
		io.WriteString(stdin, tt.stdin)
		stdin.Close()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()

		output := string(stdout.Bytes())
		if strings.Compare(output, tt.stdout) != 0 {
			t.Errorf("execution stdout (%s) => %q, want %q", tt.args, output, tt.stdout)
		}
		errors := string(stderr.Bytes())
		if strings.Compare(errors, tt.stderr) != 0 {
			t.Errorf("execution stderr (%s) => %q, want %q", tt.args, errors, tt.stderr)
		}

		e, err := getExitcode(err)
		if err != nil {
			t.Fatalf("comparing exitcode (%s) failed: %v", tt.args, err)
		}
		if e != tt.exitcode {
			t.Errorf("execution exitcode (%s) => %q, want %q", tt.args, e, tt.exitcode)
		}
	}

}

func getExitcode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return 0, fmt.Errorf("failed to cast exitcode %v", err)
	}
	waitStatus := exitError.Sys().(syscall.WaitStatus)
	return waitStatus.ExitStatus(), nil
}
