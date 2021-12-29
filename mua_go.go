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
	fmt.Println("yolo!")

	local, _ := net.ResolveTCPAddr("tcp4", "::")
	remote, _ := net.ResolveTCPAddr("tcp4", "smtp-mail.outlook.com:587")

	conn, err := net.DialTCP("tcp4", local, remote)
	handleErr(err)

	reader := bufio.NewReader(conn)

	line, err := reader.ReadBytes('\n') // look for 220
	handleErr(err)
	fmt.Println(string(line))

	// Now we say HELO
	written, err := conn.Write([]byte("EHLO\r\n")) // resolve to host IP
	fmt.Printf("written %d, err: %v\n", written, err)

	// Normally, the response to EHLO will be a multiline reply.  Each line
	// of the response contains a keyword and, optionally, one or more
	// parameters.  Following the normal syntax for multiline replies, these
	// keyworks follow the code (250) and a hyphen for all but the last
	// line, and the code and a space for the last line.  The syntax for a
	// positive response, using the ABNF notation and terminal symbols of
	// [8]
	// Read until a none STATUS- line
	welcomeResp := make([]string, 0)
	for {
		line, err = reader.ReadBytes('\n')
		handleErr(err)
		welcomeResp = append(welcomeResp, string(line))
		if line[3] != '-' {
			break
		}
	}
	fmt.Println(welcomeResp)

	conn.Write([]byte("STARTTLS\r\n"))

	line, err = reader.ReadBytes('\n') // look for 220
	fmt.Println(string(line))
	handleErr(err)

	// Say hello again as per spec with upgraded conn
	// if rootca's null, uses host os's rootca's
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: "smtp-mail.outlook.com",
	})
	written, err = tlsConn.Write([]byte("EHLO\r\n")) // resolve to host IP
	fmt.Printf("written tls %d, err: %v\n", written, err)

}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
