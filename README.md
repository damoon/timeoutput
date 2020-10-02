# timeoutput

Limits the execution time of a command.

[![Go Report Card](https://goreportcard.com/badge/github.com/damoon/timeoutput)](https://goreportcard.com/report/github.com/damoon/timeoutput)


## Usage

`timeoutput <global timeout> <output timeout> <command>...`

The maximum execution time is limited by the first parameter.  
The maximum time between outputs to stderr or stdout is limited by the second parameter.  
All following arguments are interpreted as the command to execute.

The exit code is none zero for timeouts and forwards the exit code from the executed command otherwise.

## Example

command
```
./timeoutput 10 2 bash -c "while sleep 1; do echo hello; done"
```
output
```
hello
hello
hello
hello
hello
hello
hello
hello
hello
```

## Installation

go get github.com/damoon/timeoutput
