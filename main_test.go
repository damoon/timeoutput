package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
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
	{[]string{"bash", "-c", "while sleep 1; do echo hello; done"},
		"",
		"hello\nhello\nhello\nhello\n",
		"exit status 255\n",
		11 * time.Second},
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
			cmd.Run()
			elapsed := time.Since(start)

			output := stdout.String()
			if strings.Compare(output, tt.stdout) != 0 {
				t.Errorf("execution stdout (%s) => %q, want %q", tt.args, output, tt.stdout)
			}
			errors := stderr.String()
			if strings.Compare(errors, tt.stderr) != 0 {
				t.Errorf("execution stderr (%s) => %q, want %q", tt.args, errors, tt.stderr)
			}
			if elapsed > tt.maxDuration {
				t.Errorf("execution duration (%s) => %q, want %q", tt.args, elapsed, tt.maxDuration)
			}
		})
	}
}
