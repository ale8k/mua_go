package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
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
func readLine(reader *bufio.Reader) {
	line, err := reader.ReadBytes('\n') // look for 220
	handleErr(err)
	fmt.Printf("S: %v", string(line))
}

// Handles error generic, panics if error found
func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
