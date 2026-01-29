package engine

type IntFlag uint32

const (
	WORDS IntFlag = 1 << iota
	CHARSEP
)

const (
	ColorRed         = "\033[31m"
	ColorMagenta     = "\033[35m"
	ColorBrightBlack = "\033[0;90;49m"
	ColorReset       = "\033[0m"
)

type Config struct {
	BufferSize    int // bytes
	BufferMaxSize int // bytes

	CharSep    string
	IgnoreCase bool
	ExactWord  bool
	Wildcard   bool

	NoColor     bool
	AllowBinary bool
}

func CombineFlags(flags ...IntFlag) IntFlag {
	var combined IntFlag
	for _, flag := range flags {
		combined |= flag
	}
	return combined
}
