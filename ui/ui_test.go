package ui

import (
	"testing"

	"github.com/pkg/profile"
)

func TestUI(t *testing.T) {
	defer profile.Start().Stop()
	RunUI(&FeefUI{})
}
