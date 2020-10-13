package ui

import (
	"testing"

	"github.com/pkg/profile"
)

func TestUI(t *testing.T) {
	// TODO: make this actually work headlessly (for CI)
	defer profile.Start().Stop()
	RunUI(&FeefUI{})
}
