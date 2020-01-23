package logger

import "fmt"

import "os"

var caseCount = 1

func NewError(content string) {
	fmt.Fprintf(os.Stderr, "\t\u001b[31m\u2A2F Error:\u001b[0m %s\n", content)
	os.Exit(1)
}

func NewSuccess(content string) {
	fmt.Fprintf(os.Stdout, "\t\u001b[32m\u2713 Pass: \u001b[0m %s\n", content)
}

func NewCase(content string) {
	fmt.Fprintf(os.Stdout, "Case %d: %s\n", caseCount, content)
	caseCount++
}
