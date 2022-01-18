package cmd

import (
	"fmt"
	"os"
	"strconv"
)

func isInt(in string) bool {
	_, err := strconv.ParseInt(in, 10, 64)
	return err == nil
}

func isUint(in string) bool {
	_, err := strconv.ParseUint(in, 10, 64)
	return err == nil
}

func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func println(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func noMoreThanOneTrue(bools ...bool) bool {
	var seen bool
	for _, b := range bools {
		if b {
			if seen {
				return false
			}
			seen = true
		}
	}
	return true
}

type Input struct {
	Brange BlockRange
	Cursor string
}

func checkArgs(cursor string, args []string) (out *Input, err error) {
	if !((len(args) == 0 && cursor != "") || (len(args) > 0)) {
		return nil, fmt.Errorf("expecting between 1 and 3 arguments")
	}
	out = &Input{
		Cursor: cursor,
	}
	startBlock := args[0]
	stopBlock := ""
	if len(args) > 1 {
		stopBlock = args[1]
	}
	out.Brange, err = newBlockRange(startBlock, stopBlock)
	if err != nil {
		return nil, fmt.Errorf("unable to determined block range: %w", err)
	}
	return out, nil
}
