package util

import (
	"os"
	"sync"

	"github.com/mattn/go-isatty"
)

var (
	isTTY    bool
	checkTTY sync.Once
)

// IsTTY checks if is a terminal.
func IsTTY() bool {
	checkTTY.Do(func() {
		isTTY = isatty.IsTerminal(os.Stdout.Fd())
	})

	return isTTY
}
