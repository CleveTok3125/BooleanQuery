package engine

type IntFlag uint32

const (
	WORDS IntFlag = 1 << iota
	CHARSEP
)

const (
	ColorRed         = "\033[31m"
	ColorBrightBlack = "\033[0;90;49m"
	ColorReset       = "\033[0m"
)

type Config struct {
	BufferSize    int // bytes
	BufferMaxSize int // bytes
	CharSep       string
	IgnoreCase    bool
	ExactWord     bool
	NoColor       bool
	NoIndex       bool
	NoFilename    bool
}

func CombineFlags(flags ...IntFlag) IntFlag {
	var combined IntFlag
	for _, flag := range flags {
		combined |= flag
	}
	return combined
}
