package main

import (
	"net/http"
	"time"

	"github.com/gobwas/glob"
)

type NotifyMode int

const (
	newItems NotifyMode = iota
	allItems
)

type GetParam struct {
	client     *http.Client
	timeout    time.Duration
	urls       []string
	maxThreads int
}

type NotifyParam struct {
	GetParam
	mode NotifyMode
	poll time.Duration
}

type FilterParam struct {
	max int
	// urls    []string
	sort bool // TODO: currently sorts by date. Make it clearer, maybe support sorting by other things
	item glob.Glob
}
