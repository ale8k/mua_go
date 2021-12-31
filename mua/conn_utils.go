package mua

import (
	"bufio"
	"fmt"
)

// Writes a line to the buffer safely, appends CRLF and emits immediately
func writeCRLFFlush(readWriter *bufio.ReadWriter, msg string) error {
	_, err := readWriter.WriteString(msg + "\r\n")
	if err != nil {
		return err
	}
	err = readWriter.Flush()
	if err != nil {
		return err
	}
	fmt.Printf("C: %q\n", msg)
	return nil
}

// Reads EHLO response and panics on failure
func readEhloResponse(reader *bufio.ReadWriter) ([]string, error) {
	var err error
	welcomeResp := make([]string, 0)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		welcomeResp = append(welcomeResp, string(line))
		if line[3] != '-' {
			break
		}
	}
	for _, v := range welcomeResp {
		fmt.Printf("S: %v", v)
	}
	return welcomeResp, err
}

// Reads single line and panics on failure
func readLine(reader *bufio.ReadWriter) (string, error) {
	line, err := reader.ReadBytes('\n') // look for 220
	if err != nil {
		return "", err
	}
	fmt.Printf("S: %v", string(line))
	return string(line), nil
}

// Handles error generic, panics if error found
func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
