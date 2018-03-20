package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

var mainTests = []struct {
	args        []string
	stdin       string
	stdout      string
	stderr      string
	maxDuration time.Duration
}{
	{[]string{"true"}, "", "", "", 1 * time.Second},
	{[]string{"false"}, "", "", "exit status 1\n", 1 * time.Second},
	{[]string{"echo", "-n", "Hello, world!"}, "", "Hello, world!", "", 1 * time.Second},
	{[]string{"wc", "-c"}, "Hello, world!", "13\n", "", 1 * time.Second},
	{[]string{"sleep", "5"}, "", "", "exit status 255\n", 3 * time.Second},
	{[]string{"bash", "-c", "while sleep 1; do echo hello; done"}, "", "hello\nhello\nhello\nhello\n", "exit status 255\n", 11 * time.Second},
}

func TestTrue(t *testing.T) {

	for _, tt := range mainTests {
		t.Run(fmt.Sprintf("command %s", tt.args), func(t *testing.T) {
			t.Parallel()

			args := []string{"go", "run", "main.go", "5", "2"}
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

			start := time.Now()
			err = cmd.Run()
			elapsed := time.Since(start)

			output := string(stdout.Bytes())
			if strings.Compare(output, tt.stdout) != 0 {
				t.Errorf("execution stdout (%s) => %q, want %q", tt.args, output, tt.stdout)
			}
			errors := string(stderr.Bytes())
			if strings.Compare(errors, tt.stderr) != 0 {
				t.Errorf("execution stderr (%s) => %q, want %q", tt.args, errors, tt.stderr)
			}
			if elapsed > tt.maxDuration {
				t.Errorf("execution duration (%s) => %q, want %q", tt.args, elapsed, tt.maxDuration)
			}
		})
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
