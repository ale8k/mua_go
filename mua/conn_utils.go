package mua

import (
	"bufio"
	"fmt"
)

// Writes a line to the buffer safely, appends CRLF and emits immediately, and panics on failure
func writeCRLFFlush(readWriter *bufio.ReadWriter, msg string) {
	_, err := readWriter.WriteString(msg + "\r\n")
	handleErr(err)
	err = readWriter.Flush()
	handleErr(err)
	fmt.Printf("C: %q\n", msg)
}

// Reads EHLO response and panics on failure
func readEhloResponse(reader *bufio.ReadWriter) []string {
	welcomeResp := make([]string, 0)
	for {
		line, err := reader.ReadBytes('\n')
		handleErr(err)
		welcomeResp = append(welcomeResp, string(line))
		if line[3] != '-' {
			break
		}
	}
	for _, v := range welcomeResp {
		fmt.Printf("S: %v", v)
	}
	return welcomeResp
}

// Reads single line and panics on failure
func readLine(reader *bufio.ReadWriter) string {
	line, err := reader.ReadBytes('\n') // look for 220
	handleErr(err)
	fmt.Printf("S: %v", string(line))
	return string(line)
}

// Handles error generic, panics if error found
func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
