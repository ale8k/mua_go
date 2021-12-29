package mua

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// A mail client capable of sending mail
type MailClient struct {
	// The mail server to use for resolution and sending
	mailServerAddr string

	// The address to mail from
	address string

	// The password for the address to mail from
	password string

	// The internal connection for this client
	smtpConnection net.Conn

	// The internal read/writer for this client when conducting
	// smtp transmission
	smtpReadWriter *bufio.ReadWriter
}

// Creates a new mail client
func NewMailClient(mailaddr string, password string) *MailClient {
	return &MailClient{
		mailServerAddr: "smtp-mail.outlook.com:587", // TODO: make configurable
		address:        mailaddr,
		password:       password,
	}
}

// Initiates an insecure connection to the mail server
func (mc *MailClient) connectSMTPInsecure() error {
	local, _ := net.ResolveTCPAddr("tcp4", "::")
	remote, _ := net.ResolveTCPAddr("tcp4", mc.mailServerAddr)
	conn, err := net.DialTCP("tcp4", local, remote)
	if err != nil {
		return err
	}
	handleErr(err)
	mc.smtpReadWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	readLine(mc.smtpReadWriter)               // S: Welcome
	writeCRLFFlush(mc.smtpReadWriter, "EHLO") // C: EHLO
	readEhloResponse(mc.smtpReadWriter)       // S: EHLO RESP
	mc.smtpConnection = conn
	return nil
}

// Performs a TLS negotiation using the host os's CAroot
// upgrading the internal connection to a TLS one, additionally
// updates the reader to look at the new TLS conn
func (mc *MailClient) upgradeSMTPConnectionTLS() {
	writeCRLFFlush(mc.smtpReadWriter, "STARTTLS")
	// S: 220 TLS READY TO INITIATE
	readLine(mc.smtpReadWriter)
	addr := strings.Split(mc.mailServerAddr, ":")
	tlsConn := tls.Client(mc.smtpConnection, &tls.Config{
		ServerName: addr[0],
	})
	mc.smtpConnection = tlsConn
	mc.smtpReadWriter = bufio.NewReadWriter(bufio.NewReader(tlsConn), bufio.NewWriter(tlsConn))
	fmt.Println("---TLS CONNECTION ESTABLISHED---")
	// C: EHLO (TLS)
	writeCRLFFlush(mc.smtpReadWriter, "EHLO")
	// S: EHLO RESP
	readEhloResponse(mc.smtpReadWriter)
}

// Performs a basic auth login against the client
// returns whether or not the login was succesful,
// false denotes either the user or password were incorrect
func (mc *MailClient) loginBasic() bool {
	// C: OAUTH Authenticate
	writeCRLFFlush(mc.smtpReadWriter, "AUTH LOGIN")
	usernameResp := strings.Split(readLine(mc.smtpReadWriter), " ")

	if r, err := strconv.Atoi(usernameResp[0]); r == 334 {
		// Username
		handleErr(err)
		b64Username := base64.StdEncoding.EncodeToString([]byte(mc.address))
		writeCRLFFlush(mc.smtpReadWriter, b64Username)

		// Password
		passResp := strings.Split(readLine(mc.smtpReadWriter), " ")
		if r2, err := strconv.Atoi(passResp[0]); r2 == 334 {
			handleErr(err)
			b64Password := base64.StdEncoding.EncodeToString([]byte(mc.password))
			writeCRLFFlush(mc.smtpReadWriter, b64Password)

			resLine := strings.Split(readLine(mc.smtpReadWriter), " ")
			code := resLine[0]
			if c, _ := strconv.Atoi(code); c == 235 {
				return true
			}
		}
	}
	return false
}

// Initialises / reinitialises the smtp client connection
func (mc *MailClient) initialiseSMTP() {
	mc.connectSMTPInsecure()
	mc.upgradeSMTPConnectionTLS()
	mc.loginBasic()
}

// Sends mail to the given address securely
func (mc *MailClient) SendNewMail(recipientAddress string, body string) {
	mc.initialiseSMTP()
	// Set send address
	writeCRLFFlush(mc.smtpReadWriter, fmt.Sprintf("MAIL FROM: %s", mc.address))
	readLine(mc.smtpReadWriter) // TODO: look for 250
	// Set receive address
	writeCRLFFlush(mc.smtpReadWriter, fmt.Sprintf("RCPT TO: %s", recipientAddress))
	readLine(mc.smtpReadWriter) // TODO: look for 250
	// Prep data for tranmission
	writeCRLFFlush(mc.smtpReadWriter, "DATA")
	readLine(mc.smtpReadWriter) // TODO: look for 354 (S: 354 Start mail input; end with <CRLF>.<CRLF>)
	writeCRLFFlush(mc.smtpReadWriter, body)
	readLine(mc.smtpReadWriter)
}
