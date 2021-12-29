package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type MailClient struct {
	// The mail server to use for resolution and sending
	mailServerAddr string

	// The address to mail from
	address string

	// The password for the address to mail from
	password string

	// The internal connection for this client
	connection net.Conn

	// The internal reader for this client
	reader *bufio.Reader
}

func NewMailClient(mailaddr string, password string) *MailClient {
	client := &MailClient{
		mailServerAddr: "smtp-mail.outlook.com:587",
		address:        mailaddr,
		password:       password,
	}
	client.connectInsecure()
	return client
}

// Initiates an insecure connection to the mail server
func (mc *MailClient) connectInsecure() error {
	local, _ := net.ResolveTCPAddr("tcp4", "::")
	remote, _ := net.ResolveTCPAddr("tcp4", mc.mailServerAddr)
	conn, err := net.DialTCP("tcp4", local, remote)
	if err != nil {
		return err
	}
	handleErr(err)
	mc.reader = bufio.NewReader(conn)
	readLine(mc.reader)         // S: Welcome
	writeLine(conn, "EHLO\r\n") // C: EHLO
	readEhloResponse(mc.reader) // S: EHLO RESP
	mc.connection = conn
	return nil
}

// Performs a TLS negotiation using the host os's CAroot
// upgrading the internal connection to a TLS one, additionally
// updates the reader to look at the new TLS conn
func (mc *MailClient) UpgradeConnectionTLS() {
	conn, _ := mc.connection.(*net.TCPConn)
	writeLine(conn, "STARTTLS\r\n")
	// S: 220 TLS READY TO INITIATE
	readLine(mc.reader)
	addr := strings.Split(mc.mailServerAddr, ":")
	tlsConn := tls.Client(mc.connection, &tls.Config{
		ServerName: addr[0],
	})
	mc.connection = tlsConn
	mc.reader = bufio.NewReader(tlsConn)
	fmt.Println("---TLS CONNECTION ESTABLISHED---")
	// C: EHLO (TLS)
	writeLineTLS(tlsConn, "EHLO\r\n")
	// S: EHLO RESP
	readEhloResponse(mc.reader)
}

// Performs a basic auth login against the client
// returns whether or not the login was succesful,
// false denotes either the user or password were incorrect
func (mc *MailClient) LoginBasicSecure() bool {
	tlsConn, ok := mc.connection.(*tls.Conn)
	if !ok {
		panic("connection is not secure! upgrade connection first via UpgradeConnecTLS()")
	}
	// C: OAUTH Authenticate
	writeLineTLS(tlsConn, "AUTH LOGIN\r\n")
	usernameResp := strings.Split(readLine(mc.reader), " ")

	if r, err := strconv.Atoi(usernameResp[0]); r == 334 {
		decode64 := base64.StdEncoding.DecodeString

		// Username
		handleErr(err)
		b64Username := base64.StdEncoding.EncodeToString([]byte(mc.address))
		writeLineTLS(tlsConn, b64Username+"\r\n")

		// Password
		passResp := strings.Split(readLine(mc.reader), " ")
		if r2, err := strconv.Atoi(passResp[0]); r2 == 334 {
			handleErr(err)
			passReq, err := decode64(passResp[1])
			handleErr(err)
			fmt.Println(string(passReq))

			b64Password := base64.StdEncoding.EncodeToString([]byte(mc.password))
			writeLineTLS(tlsConn, b64Password+"\r\n")

			resLine := strings.Split(readLine(mc.reader), " ")
			code := resLine[0]
			if c, _ := strconv.Atoi(code); c == 235 {
				return true
			}
		}
	}
	return false
}

// Sends mail to the given address
func (mc *MailClient) SendNewMail(recipientAddress string, body string) {
	tlsConn, ok := mc.connection.(*tls.Conn)
	if !ok {
		panic("connection is unsecure! upgrade connection!")
	}
	// Set send address
	writeLineTLS(tlsConn, fmt.Sprintf("MAIL FROM: %s\r\n", mc.address))
	readLine(mc.reader) // TODO: look for 250

	// Set receive address
	writeLineTLS(tlsConn, fmt.Sprintf("RCPT TO: %s\r\n", recipientAddress))
	readLine(mc.reader) // TODO: look for 250

	// Prep data for tranmission
	writeLineTLS(tlsConn, "DATA\r\n")
	readLine(mc.reader) // TODO: look for 354 (S: 354 Start mail input; end with <CRLF>.<CRLF>)

	writeLineTLS(tlsConn, body)
	readLine(mc.reader)
}

func main() {
	mailAddress := "bob@bob.com"
	mailPwd := ""
	client := NewMailClient(mailAddress, mailPwd)
	client.UpgradeConnectionTLS()

	if succesfulLogin := client.LoginBasicSecure(); succesfulLogin {
		fmt.Println("login?: ", succesfulLogin)
	}

	mailBuilder := MailBuilder{}
	mailBuilder.SetTo(mailAddress)
	mailBuilder.SetFrom("digletti", mailAddress)
	mailBuilder.SetSubject("testing this thing out")
	mailBuilder.UpdateMailBodyString("sending mail")

	client.SendNewMail(mailAddress, string(mailBuilder.Build()))
}
