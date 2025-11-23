package dbg

import (
	"fmt"
	"os"

	"github.com/jwalton/go-supportscolor"
)

const (
	prefix = "CAPYCHAIN"

	red   = "\033[1;31m"
	green = "\033[32m"
	reset = "\033[0m"

	prefix_color = "\033[33m"
)

func getPrefixTag() string {
	if supportscolor.Stdout().SupportsColor {
		return fmt.Sprintf("%s[%s]%s", prefix_color, prefix, reset)
	}
	return fmt.Sprintf("[%s]", prefix)
}

func getInfoTag() string {
	if supportscolor.Stdout().SupportsColor {
		return fmt.Sprintf("%s[%s]%s", green, "INFO", reset)
	}
	return "[INFO]"
}

func getErrorTag() string {
	if supportscolor.Stdout().SupportsColor {
		return fmt.Sprintf("%s[%s]%s", red, "ERROR", reset)
	}
	return "[ERROR]"
}

func Infof(format string, args ...any) {
	fmt.Printf(
		"%s %s %s\n",
		getPrefixTag(),
		getInfoTag(),
		fmt.Sprintf(format, args...),
	)
	os.Stdout.Sync()
}

func Errorf(format string, args ...any) {
	fmt.Printf(
		"%s %s %s\n",
		getPrefixTag(),
		getErrorTag(),
		fmt.Sprintf(format, args...),
	)
	os.Stdout.Sync()
}

func ExitWithErrorf(format string, args ...any) {
	fmt.Printf(
		"%s %s %s\n",
		getPrefixTag(),
		getErrorTag(),
		fmt.Sprintf(format, args...),
	)
	os.Stdout.Sync()
	os.Exit(1)
}
