package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"BooleanQuery/engine"
)

func collectFiles(root string, followSymlinks bool, expanded *[]string, visited map[string]bool, noWarn bool) {
	var walk func(string)
	walk = func(currentPath string) {
		info, err := os.Lstat(currentPath)
		if err != nil {
			if !noWarn {
				fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", currentPath, err)
			}
			return
		}

		targetInfo := info

		if info.Mode()&os.ModeSymlink != 0 {
			if !followSymlinks {
				return
			}

			resolvedPath, err := filepath.EvalSymlinks(currentPath)
			if err != nil {
				if !noWarn {
					fmt.Fprintf(os.Stderr, "Error resolving symlink %s: %v\n", currentPath, err)
				}
				return
			}
			resolvedInfo, err := os.Stat(resolvedPath)
			if err != nil {
				return
			}
			targetInfo = resolvedInfo

			absPath, _ := filepath.Abs(resolvedPath)
			if visited[absPath] {
				if !noWarn {
					fmt.Fprintf(os.Stderr, "bq: warning: recursive symlink loop detected at %s\n", currentPath)
				}
				return
			}
			visited[absPath] = true
		}

		if targetInfo.IsDir() {
			entries, err := os.ReadDir(currentPath)
			if err != nil {
				if !noWarn {
					fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", currentPath, err)
				}
				return
			}
			for _, entry := range entries {
				walk(filepath.Join(currentPath, entry.Name()))
			}
		} else {
			if targetInfo.Mode().IsRegular() {
				*expanded = append(*expanded, currentPath)
			}
		}
	}
	walk(root)
}

func Run(e *engine.Engine) bool {
	watchTerminalSize()

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

	isRecursive := cli.Recursive || cli.DereferenceRecursive
	if isRecursive {
		cli.ShowFilePrefix = true
	}

	if len(cli.Files) == 0 {
		if isRecursive {
			cli.Files = []string{"."}
		} else {
			stdoutWriter := bufio.NewWriter(os.Stdout)
			defer stdoutWriter.Flush()
			return processInput(stdoutWriter, e, os.Stdin, "")
		}
	}

	var expandedFiles []string
	visitedSymlinks := make(map[string]bool)

	for _, path := range cli.Files {
		info, err := os.Stat(path)
		if err != nil {
			if !cli.NoWarn {
				fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
			}
			continue
		}

		if info.IsDir() {
			if isRecursive {
				collectFiles(path, cli.DereferenceRecursive, &expandedFiles, visitedSymlinks, cli.NoWarn)
			} else {
				if !cli.NoWarn {
					fmt.Fprintf(os.Stderr, "bq: %s: Is a directory\n", path)
				}
			}
		} else {
			expandedFiles = append(expandedFiles, path)
		}
	}

	cli.Files = expandedFiles

	if len(cli.Files) == 0 {
		return false
	}

	stdoutWriter := bufio.NewWriter(os.Stdout)
	defer stdoutWriter.Flush()

	globalFound := false

	if len(cli.Files) <= 1 || cli.Stream {
		if len(cli.Files) == 0 {
			return processInput(stdoutWriter, e, os.Stdin, "")
		}

		for _, path := range cli.Files {
			found, err := processFile(stdoutWriter, e, path)
			if err != nil {
				stdoutWriter.Flush()
				fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
			}
			if found {
				globalFound = true
			}
			stdoutWriter.Flush()
		}
		return globalFound
	}

	numCPU := runtime.NumCPU()
	sem := make(chan struct{}, numCPU)
	var wg sync.WaitGroup

	tempResults := make([]string, len(cli.Files))

	var foundLock sync.Mutex

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

				foundLock.Lock()
				globalFound = true
				foundLock.Unlock()
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

	return globalFound
}
