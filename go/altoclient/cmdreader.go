package main

import (
	"bufio"
	"io"
	"strings"
	"os"
	"fmt"
	)
	
type CmdReader struct {
	Prompt string
	WPrompt io.Writer
	scan *bufio.Scanner
	closed bool
}

func NewCmdReader(r io.Reader) CmdReader {
	if r == nil {
		r = os.Stdin
	}
	return CmdReader{scan: bufio.NewScanner(r),
					 Prompt: "* ",
					 WPrompt: os.Stdout,
					 closed: false,
					 }
}

func (this *CmdReader) NextCmd() []string {
	if this.closed {
		return nil
	}
	if this.Prompt != "" && this.WPrompt != nil {
		fmt.Fprint(this.WPrompt, this.Prompt)
		flushWriter(this.WPrompt)
	}
	var line string
	for {
		if !this.scan.Scan() {
			this.closed = true
			break
		}
		s := this.scan.Text()
		if strings.HasSuffix(s, "\\") {
			line += s[0:len(s)-1]
		} else {
			line += s
			break
		}
	}
	if line == "" && this.closed {
		return nil
	} else {
		return strings.Fields(line)
	}
}

func flushWriter(w io.Writer) {
	switch v := w.(type) {
	case *bufio.Writer:
		v.Flush()
	}
}
	
