package mua

import (
	"fmt"
	"strconv"
	"strings"
)

// Individual status handlers for particular smtp resp codes

// Splits the line and status, returning the parsed code and status line
func getCodeAndStatus(line string) (int, string) {
	code := strings.TrimSpace(line[:3])
	parsedCode, _ := strconv.Atoi(code)
	statusLine := line[3:]
	return parsedCode, statusLine
}

func handleSmtpResponse(line string, expectedCode int) (bool, int, string) {
	code, statusLine := getCodeAndStatus(line)

	switch code {
	case 220: // Service ready
		fmt.Println("all ok")
	case 221: // Service closing
	case 250: // Action successful
	case 334: // User/pass auth requested
	case 354: // Begin message
	case 421: // Service unvailable and connection closing
	case 450: // Users mailbox unavailable
	case 451: // Command aborted, internal server error
	case 452: // Command aborted, internal server error, no storage, usually overload from us
	case 455: // Service temporarily unavaialble
	case 500: // Unrecognised command
	case 501: // Syntax error, usually mail address
	case 502: // Unimplemented
	case 503: // Missing mail command, can happen when not auth too, order of commands wrong
	case 521: // Host doesn't accept mail
	case 535: // Auth failed
	case 541: // Couldn't be delivered, usually spam filter
	}

	if code != expectedCode {
		return false, code, statusLine
	}
	return true, code, statusLine
}
