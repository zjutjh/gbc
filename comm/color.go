package comm

import (
	"fmt"
	"io"
	"os"
)

const (
	Black   = "\033[1:30m"
	Red     = "\033[1;31m"
	Green   = "\033[1;32m"
	Yellow  = "\033[1;33m"
	Blue    = "\033[1;34m"
	Magenta = "\033[1;35m"
	Cyan    = "\033[1;36m"
	White   = "\033[1;37m"

	Reset = "\033[0m"
)

const (
	Debug = "\u001B[3;34m"
	Info  = "\u001B[22;37m"
	Error = "\u001B[7;31m"
	// Look          = "\u001B[4;32m"
	Look          = "\u001B[22;32m"
	UserInterface = "\u001B[4;37m"
)

func Fprintf(w io.Writer, c, format string, a ...any) {
	format = fmt.Sprintf("%s %s %s%s", c, format, Reset, NewLine)
	fmt.Fprintf(w, format, a...)
}

func OutputDebug(format string, a ...any) {
	Fprintf(os.Stdout, Debug, format, a...)
}

func OutputInfo(format string, a ...any) {
	Fprintf(os.Stdout, Info, format, a...)
}

func OutputError(format string, a ...any) {
	Fprintf(os.Stderr, Error, format, a...)
}

func OutputLook(format string, a ...any) {
	Fprintf(os.Stdout, Look, format, a...)
}

func OutputUI(w io.Writer, format string, a ...any) {
	Fprintf(w, UserInterface, format, a...)
}
