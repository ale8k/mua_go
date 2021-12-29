package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// SMTP Server	smtp-mail.outlook.com
// Username	Your full Outlook.com email address
// Password	Your Outlook.com password
// SMTP Port	587
// SMTP TLS/SSL Encryption Required	Yes

func main() {
	local, _ := net.ResolveTCPAddr("tcp4", "::")
	remote, _ := net.ResolveTCPAddr("tcp4", "smtp-mail.outlook.com:587")
	fmt.Println("---STARTING SMTP CONNECTION---")
	conn, err := net.DialTCP("tcp4", local, remote)
	handleErr(err)
	reader := bufio.NewReader(conn)
	inputReader := bufio.NewReader(os.Stdin)

	// S: WELCOME
	readLine(reader)

	// C: EHLO
	writeLine(conn, "EHLO\r\n")

	// S: EHLO RESP
	readEhloResponse(reader)

	// C: STARTTLS
	writeLine(conn, "STARTTLS\r\n")

	// S: 220 TLS READY TO INITIATE
	readLine(reader)

	// Client upgrade connection
	tlsConn, reader := upgradeConnectionTLS(conn, "smtp-mail.outlook.com")

	// C: EHLO (TLS)
	writeLineTLS(tlsConn, "EHLO\r\n")

	// S: EHLO RESP
	readEhloResponse(reader)

	// C: OAUTH Authenticate
	writeLineTLS(tlsConn, "AUTH LOGIN\r\n")
	// S: Read oauth resp
	usernameResp := strings.Split(readLine(reader), " ")

	if r, err := strconv.Atoi(usernameResp[0]); r == 334 {
		decode64 := base64.StdEncoding.DecodeString

		// Username
		handleErr(err)
		usernameReq, err := decode64(usernameResp[1])
		handleErr(err)
		// TODO: run email regex
		fmt.Println(string(usernameReq))
		username, _, _ := inputReader.ReadLine()
		b64Username := base64.StdEncoding.EncodeToString(username)
		writeLineTLS(tlsConn, b64Username+"\r\n")

		// Password
		passResp := strings.Split(readLine(reader), " ")
		if r2, err := strconv.Atoi(passResp[0]); r2 == 334 {
			handleErr(err)
			passReq, err := decode64(passResp[1])
			handleErr(err)
			fmt.Println(string(passReq))
			password, _ := terminal.ReadPassword(0)
			b64Password := base64.StdEncoding.EncodeToString(password)

		}
	}

}

// upgrades connection to tls using host ca's
// additionally returns a buffered reader for convenience
func upgradeConnectionTLS(conn *net.TCPConn, serverName string) (*tls.Conn, *bufio.Reader) {
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: serverName,
	})
	return tlsConn, bufio.NewReader(tlsConn)
}

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
