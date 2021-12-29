package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
)

// Writes a none tls line error handled to server
// panics if write fails
// prints the line
func writeLine(conn *net.TCPConn, msg string) {
	_, err := conn.Write([]byte(msg))
	handleErr(err)
	fmt.Printf("C: %q\n", msg)
}

// Writes a tls line error handled to server
// panics if write fails
// prints the line
func writeLineTLS(conn *tls.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	handleErr(err)
	fmt.Printf("C: %q\n", msg)
}

// reads a none-tls ehlo response
func readEhloResponse(reader *bufio.Reader) []string {
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

// reads single line safely from bufio.Reader
func readLine(reader *bufio.Reader) string {
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
