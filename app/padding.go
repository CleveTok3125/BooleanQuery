package app

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

var (
	padLine int = 4
	padCol  int = 2
)

func watchTerminalSize() {
	update := func() {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return
		}

		if width < 60 {
			padLine, padCol = 0, 0
		} else if width < 120 {
			padLine, padCol = 4, 2
		} else {
			padLine, padCol = 6, 3
		}
	}

	update()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)

	go func() {
		for range sigChan {
			update()
		}
	}()
}
