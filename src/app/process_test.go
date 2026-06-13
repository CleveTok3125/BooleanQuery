package app

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"BooleanQuery/src/engine"
)

func TestPrintLine(t *testing.T) {
	tests := []struct {
		name           string
		setup          func()
		text           string
		matches        [][2]int
		index          int
		cachedPrefix   []byte
		wantContain    string
		wantNotContain string
	}{
		{
			name: "simple print with index",
			setup: func() {
				padLine, padCol = 4, 2
			},
			text:        "hello world",
			matches:     [][2]int{{0, 5}},
			index:       0,
			wantContain: "hello",
		},
		{
			name: "print with file prefix",
			setup: func() {
				padLine, padCol = 4, 2
			},
			text:         "test line",
			matches:      nil,
			index:        5,
			cachedPrefix: []byte("file.txt:"),
			wantContain:  "file.txt:",
		},
		{
			name: "no index mode",
			setup: func() {
				padLine, padCol = 4, 2
				cli.NoIndex = true
			},
			text:           "raw line",
			matches:        nil,
			index:          0,
			wantNotContain: "1",
		},
		{
			name: "with color no matches",
			setup: func() {
				padLine, padCol = 4, 2
			},
			text:        "plain text",
			matches:     nil,
			index:       3,
			wantContain: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldCli := cli
			defer func() { cli = oldCli }()
			setCLIDefaults()

			tt.setup()

			e := engine.New()
			e.Config.NoColor = true

			var buf bytes.Buffer
			var printBuf bytes.Buffer
			err := printLine(&buf, e, tt.index, tt.matches, tt.text, tt.cachedPrefix, &printBuf)
			if err != nil {
				t.Fatalf("printLine returned error: %v", err)
			}

			output := buf.String()
			if tt.wantContain != "" && !strings.Contains(output, tt.wantContain) {
				t.Errorf("output %q should contain %q", output, tt.wantContain)
			}
			if tt.wantNotContain != "" && strings.Contains(output, tt.wantNotContain) {
				t.Errorf("output %q should NOT contain %q", output, tt.wantNotContain)
			}
		})
	}
}

func TestProcessInput_Basic(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	if err := e.SetSearchTerm("error"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	var buf bytes.Buffer
	input := strings.NewReader("hello world\nerror occurred\nfine line\nerror again\n")

	found := processInput(&buf, e, input, "")
	if !found {
		t.Error("expected found=true, got false")
	}

	output := buf.String()
	if !strings.Contains(output, "error occurred") {
		t.Errorf("output should contain 'error occurred', got: %s", output)
	}
	if !strings.Contains(output, "error again") {
		t.Errorf("output should contain 'error again', got: %s", output)
	}
}

func TestProcessInput_NoMatch(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	if err := e.SetSearchTerm("nonexistent"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	var buf bytes.Buffer
	input := strings.NewReader("hello world\nnothing here\n")

	found := processInput(&buf, e, input, "")
	if found {
		t.Error("expected found=false, got true")
	}
	if buf.Len() > 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}

func TestProcessInput_CountMode(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	if err := e.SetSearchTerm("error"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	var buf bytes.Buffer
	input := strings.NewReader("error one\ngood\nerror two\nok\nerror three\n")
	cli.Count = true

	found := processInput(&buf, e, input, "")
	if !found {
		t.Error("expected found=true")
	}
	if !strings.Contains(buf.String(), "3") {
		t.Errorf("output should contain count '3', got: %s", buf.String())
	}
}

func TestProcessInput_FilesWithMatches(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	if err := e.SetSearchTerm("error"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	var buf bytes.Buffer
	input := strings.NewReader("error found\nok\n")
	cli.FilesWithMatches = true

	found := processInput(&buf, e, input, "test.txt")
	if !found {
		t.Error("expected found=true")
	}
	if !strings.Contains(buf.String(), "test.txt") {
		t.Errorf("output should contain filename, got: %s", buf.String())
	}
}

func TestProcessInput_BinaryFile(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024
	e.Config.AllowBinary = false

	input := strings.NewReader("hello\x00world\n")

	var buf bytes.Buffer

	found := processInput(&buf, e, input, "test.bin")
	if found {
		t.Error("expected found=false for binary file")
	}
}

func TestProcessFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.txt"
	if err := os.WriteFile(path, []byte("hello world\nerror found\n"), 0644); err != nil {
		t.Fatal(err)
	}

	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024
	if err := e.SetSearchTerm("error"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	var buf bytes.Buffer
	found, err := processFile(&buf, e, path)
	if err != nil {
		t.Fatalf("processFile error: %v", err)
	}
	if !found {
		t.Error("expected found=true")
	}
	if !strings.Contains(buf.String(), "error found") {
		t.Errorf("output should contain match, got: %s", buf.String())
	}
}

func TestProcessFile_NotFound(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	var buf bytes.Buffer
	_, err := processFile(&buf, e, "/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestProcessFileToTemp(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.txt"
	if err := os.WriteFile(path, []byte("hello world\nerror found\nerror again\n"), 0644); err != nil {
		t.Fatal(err)
	}

	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024
	if err := e.SetSearchTerm("error"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	tmpPath, hasContent, err := processFileToTemp(e, path)
	if err != nil {
		t.Fatalf("processFileToTemp error: %v", err)
	}
	defer os.Remove(tmpPath)
	if !hasContent {
		t.Error("expected hasContent=true")
	}
	if tmpPath == "" {
		t.Fatal("expected non-empty temp path")
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "error found") {
		t.Errorf("temp file should contain match, got: %s", data)
	}
}

func TestProcessFileToTemp_NoMatch(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.txt"
	if err := os.WriteFile(path, []byte("hello world\nall good\n"), 0644); err != nil {
		t.Fatal(err)
	}

	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024
	if err := e.SetSearchTerm("nonexistent"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	tmpPath, hasContent, err := processFileToTemp(e, path)
	if err != nil {
		t.Fatalf("processFileToTemp error: %v", err)
	}
	if hasContent {
		t.Error("expected hasContent=false for no matches")
	}
	if tmpPath != "" {
		os.Remove(tmpPath)
	}
}

func TestProcessInput_NegateQuery(t *testing.T) {
	oldCli := cli
	defer func() { cli = oldCli }()
	setCLIDefaults()
	padLine, padCol = 4, 2

	e := engine.New()
	e.Config.NoColor = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	if err := e.SetSearchTerm("error -timeout"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	var buf bytes.Buffer
	input := strings.NewReader("error occurred\nerror with timeout\njust error\n")
	found := processInput(&buf, e, input, "")
	if !found {
		t.Error("expected found=true")
	}
	output := buf.String()
	if strings.Contains(output, "timeout") {
		t.Errorf("output should not contain 'timeout' lines, got: %s", output)
	}
}
