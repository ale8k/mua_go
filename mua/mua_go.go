package mua

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	INITIAL_CONNECTION_FAILURE = "initial connection failed, response code: %d, response message: %s"
	WRITE_FAILURE              = "could not write to server, exiting... see: %w"
	EHLO_ERROR                 = "client received invalid/erroneous/unexpected EHLO response with code %d, see: %s"
)

// A mail client capable of sending mail
//
// Note: this should not be instantiated directly! Instead use mua.NewMailClient()
type MailClient struct {
	// The address to mail from
	address string

	// The password for the address to mail from
	password string

	// The mail server to use for resolution and sending
	smtpMailServerAddr string

	// The internal smtp connection for this client
	smtpConnection net.Conn

	// The internal read/writer for this client when conducting
	// smtp transmission
	smtpReadWriter *bufio.ReadWriter

	// The internal imap connection for this client
	// todo
}

// Creates a new mail client
func NewMailClient(mailaddr string, password string, smtpAddr string) *MailClient {
	return &MailClient{
		address:            mailaddr,
		password:           password,
		smtpMailServerAddr: smtpAddr,
	}
}

// Initiates an insecure connection to the mail server
func (mc *MailClient) connectSMTPInsecure() error {
	local, err := net.ResolveTCPAddr("tcp4", "::")
	remote, err := net.ResolveTCPAddr("tcp4", mc.smtpMailServerAddr)
	conn, err := net.DialTCP("tcp4", local, remote)
	if err != nil {
		return err
	}

	mc.smtpReadWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Initiate
	if ok, code, statusLine := handleSmtpResponse(readLine(mc.smtpReadWriter), 220); !ok {
		return errors.New(fmt.Sprintf(INITIAL_CONNECTION_FAILURE, code, statusLine))
	}
	// Say EHLO
	if err = writeCRLFFlush(mc.smtpReadWriter, "EHLO"); err != nil {
		return fmt.Errorf(WRITE_FAILURE, err)
	}
	// Read EHLO
	statusLines, err := readEhloResponse(mc.smtpReadWriter) // S: EHLO RESP
	if err != nil {
		return err
	}
	// Handle read EHLO
	for _, ehloLine := range statusLines {
		ok, code, statusLine := handleSmtpResponse(ehloLine, 250) // we expect 250 for each ehlo line as per spec
		if !ok {
			return errors.New(fmt.Sprintf(EHLO_ERROR, code, statusLine))
		}
	}
	// Store connection
	mc.smtpConnection = conn
	return err
}

// Performs a TLS negotiation using the host os's CAroot
// upgrading the internal connection to a TLS one, additionally
// updates the reader to look at the new TLS conn
// Only use if server complies to this extension
func (mc *MailClient) upgradeSMTPConnectionTLS() error {
	var err error

	if err = writeCRLFFlush(mc.smtpReadWriter, "STARTTLS"); err != nil {
		return fmt.Errorf(WRITE_FAILURE, err)
	}
	// S: 220 TLS READY TO INITIATE
	readLine(mc.smtpReadWriter)
	addr := strings.Split(mc.smtpMailServerAddr, ":")
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
	return nil
}

// Performs a basic auth login against the client
// returns whether or not the login was succesful,
// false denotes either the user or password were incorrect
// Only use if server complies to this extension
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

// Opens a new smtp connection, if tls is wanted, optionally turn tls on
// attempts to login with basic auth from the initial credential provided
// when creating the client
func (mc *MailClient) OpenSMTPConnection(tls bool) error {
	var err error
	if err = mc.connectSMTPInsecure(); err != nil {
		return err
	}
	if tls {
		if err = mc.upgradeSMTPConnectionTLS(); err != nil {
			return err
		}
	}
	mc.loginBasic()
	return err
}

// Attempts to close the smtp connection, if it isn't able,
// will return err else nil
func (mc *MailClient) CloseSMTPConnection() error {
	return mc.smtpConnection.Close()
}

// Sends mail to the given address securely
func (mc *MailClient) SendNewMail(recipientAddress string, body string) {
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

	// Reset mail
	writeCRLFFlush(mc.smtpReadWriter, "RSET")
	readLine(mc.smtpReadWriter)
}
