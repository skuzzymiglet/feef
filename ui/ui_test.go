package ui

import (
	"testing"

	"github.com/pkg/profile"
)

func TestUI(t *testing.T) {
	defer profile.Start().Stop()
	done := make(chan struct{}, 1)
	RunUI(&FeefUI{DoneChan: done}, done)
}
