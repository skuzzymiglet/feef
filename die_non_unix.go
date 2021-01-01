// +build windows plan9

package main

import "os"

var die []os.Signal = []os.Signal{os.Kill, os.Interrupt}
