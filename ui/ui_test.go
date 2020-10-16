package ui

import (
	"testing"

	"github.com/pkg/profile"
)

func TestUI(t *testing.T) {
	// TODO: make this actually work headlessly (for CI)
	defer profile.Start().Stop()
	done := make(chan struct{}, 1)
	RunUI(&FeefUI{DoneChan: done}, done)
}
