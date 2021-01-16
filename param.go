package main

import (
	"time"

	"github.com/antonmedv/expr/vm"
)

type NotifyMode int

const (
	newItems NotifyMode = iota + 1
	allItems
)

type GetParam struct { // this is so bad
	Fetcher
	urls []string
}

type NotifyParam struct {
	GetParam
	mode NotifyMode
	poll time.Duration
}

type FilterParam struct {
	max int
	// urls    []string
	sort       bool // TODO: currently sorts by date. Make it clearer, maybe support sorting by other things
	matcher    *vm.Program
	matcherEnv interface{} // Unused
}
