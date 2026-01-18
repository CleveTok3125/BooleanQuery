package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"BooleanQuery/engine"
)

func Run(e *engine.Engine) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	var tempFiles []string
	var tempFilesLock sync.Mutex

	go func() {
		<-c
		tempFilesLock.Lock()
		for _, f := range tempFiles {
			os.Remove(f)
		}
		tempFilesLock.Unlock()
		os.Exit(1)
	}()

	stdoutWriter := bufio.NewWriter(os.Stdout)
	defer stdoutWriter.Flush()

	if len(cli.Files) <= 1 || cli.Stream {
		if len(cli.Files) == 0 {
			processInput(stdoutWriter, e, os.Stdin, "")
			return
		}

		for _, path := range cli.Files {
			if err := processFile(stdoutWriter, e, path); err != nil {
				stdoutWriter.Flush()
				fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
			}
			stdoutWriter.Flush()
		}
		return
	}

	numCPU := runtime.NumCPU()
	sem := make(chan struct{}, numCPU)
	var wg sync.WaitGroup

	tempResults := make([]string, len(cli.Files))

	for i, path := range cli.Files {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			tmpPath, hasContent, err := processFileToTemp(e, p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", p, err)
				return
			}

			tempFilesLock.Lock()
			tempFiles = append(tempFiles, tmpPath)
			tempFilesLock.Unlock()

			if hasContent {
				tempResults[idx] = tmpPath
			} else {
				os.Remove(tmpPath)
				tempResults[idx] = ""
			}
		}(i, path)
	}

	wg.Wait()

	for _, tmpPath := range tempResults {
		if tmpPath == "" {
			continue
		}

		f, err := os.Open(tmpPath)
		if err == nil {
			io.Copy(stdoutWriter, f)
			f.Close()
		}

		os.Remove(tmpPath)
		stdoutWriter.Flush()
	}
}
