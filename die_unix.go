// +build !windows !plan9

package main

import (
	"os"
	"syscall"
)

var die []os.Signal = []os.Signal{
	syscall.SIGPIPE, os.Interrupt, os.Kill,
}
