package mua

/*
A mail client implemented in Go, see examples for usage

This package reports SMTP server specific errors directly back to user
when encountered, as such expect all errors to either a) be SMTP specific
or b) bespoke, such as a generic connection failure, regex failure etc.

Currently supports the following SMTP extensions:
	- STARTTLS
	- AUTH (LOGIN)
*/
import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	INITIAL_CONNECTION_FAILURE = "initial connection failed, response code: %d, response message: %s"
	WRITE_FAILURE              = "could not write to server, exiting... see: %w"
	READ_FAILURE               = "could not read from server, exiting... see: %w"
	EHLO_ERROR                 = "client received invalid/erroneous/unexpected EHLO response with code %d, see: %s"
	AUTH_LOGIN_ERROR           = "could not login, response code: %d, response message: %s"
	MAIL_FROM_ERROR            = "MAIL FROM command rejected, see code: %d, response message: %s"
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
	smtpConnection net.Conn // TODO: Buffer connections for concurrent writing or open many

	// Supported SMTP extensions
	smtpExtensionsSupported []string // TODO: append to this list the supported extensions when read
	// if we attempt one that doesn't exist, err

	// The internal read/writer for this client when conducting
	// smtp transmission
	smtpReadWriter *bufio.ReadWriter

	// The internal imap connection for this client
	// todo
}

// Creates a new mail client for SENDING
func NewMailClient(mailaddr string, password string, smtpAddr string) *MailClient {
	return &MailClient{
		address:            mailaddr,
		password:           password,
		smtpMailServerAddr: smtpAddr,
	}
}

// Initiates an insecure connection to the mail server
func (mc *MailClient) connectSMTPInsecure() error {
	// Open conn
	local, err := net.ResolveTCPAddr("tcp4", "::")
	remote, err := net.ResolveTCPAddr("tcp4", mc.smtpMailServerAddr)
	conn, err := net.DialTCP("tcp4", local, remote)
	if err != nil {
		return err
	}
	// Store buffered readwriter
	mc.smtpReadWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	// Initiate
	line, err := readLine(mc.smtpReadWriter)
	if err != nil {
		return fmt.Errorf(READ_FAILURE, err)
	}
	// Handle resp
	if ok, code, statusLine := handleSmtpResponse(line, 220); !ok {
		return errors.New(fmt.Sprintf(INITIAL_CONNECTION_FAILURE, code, statusLine))
	}
	// Say EHLO
	if err = writeCRLFFlush(mc.smtpReadWriter, "EHLO"); err != nil {
		return fmt.Errorf(WRITE_FAILURE, err)
	}
	// Read EHLO
	statusLines, err := readEhloResponse(mc.smtpReadWriter) // S: EHLO RESP
	if err != nil {
		return fmt.Errorf(READ_FAILURE, err)
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
	// TODO: Check if this extension is supported
	var err error
	// Submit extension request
	if err = writeCRLFFlush(mc.smtpReadWriter, "STARTTLS"); err != nil {
		return fmt.Errorf(WRITE_FAILURE, err)
	}
	// S: 220 TLS READY TO INITIATE
	line, err := readLine(mc.smtpReadWriter)
	if err != nil {
		return fmt.Errorf(READ_FAILURE, err)
	}
	// Handle resp
	if ok, code, statusLine := handleSmtpResponse(line, 220); !ok {
		return errors.New(fmt.Sprintf(INITIAL_CONNECTION_FAILURE, code, statusLine))
	}
	// Get domain only part
	addr := strings.Split(mc.smtpMailServerAddr, ":")
	// Wrap connection
	tlsConn := tls.Client(mc.smtpConnection, &tls.Config{
		ServerName: addr[0],
	})
	// Update connection
	mc.smtpConnection = tlsConn
	// Update buffered readwriter
	mc.smtpReadWriter = bufio.NewReadWriter(bufio.NewReader(tlsConn), bufio.NewWriter(tlsConn))
	// Debug log
	fmt.Println("---TLS CONNECTION ESTABLISHED---")
	// C: EHLO (TLS)
	if err = writeCRLFFlush(mc.smtpReadWriter, "EHLO"); err != nil {
		return fmt.Errorf(WRITE_FAILURE, err)
	}
	// S: EHLO RESP
	statusLines, err := readEhloResponse(mc.smtpReadWriter)
	if err != nil {
		return fmt.Errorf(READ_FAILURE, err)
	}
	// Handle read EHLO
	for _, ehloLine := range statusLines {
		ok, code, statusLine := handleSmtpResponse(ehloLine, 250) // we expect 250 for each ehlo line as per spec
		if !ok {
			return errors.New(fmt.Sprintf(EHLO_ERROR, code, statusLine))
		}
	}
	return err
}

// Performs a basic auth login against the client
// returns whether or not the login was succesful,
// false denotes either the user or password were incorrect
// Only use if server complies to this extension
func (mc *MailClient) loginBasic() (bool, error) {
	// TODO: Check if this extension is supported
	var err error
	// C: Authenticate basic login
	if err = writeCRLFFlush(mc.smtpReadWriter, "AUTH LOGIN"); err != nil {
		return false, fmt.Errorf(WRITE_FAILURE, err)
	}
	// Read 334 (username)
	line, err := readLine(mc.smtpReadWriter)
	if err != nil {
		return false, fmt.Errorf(READ_FAILURE, err)
	}
	// Handle resp
	ok, code, statusLine := handleSmtpResponse(line, 334)
	// Error out on a none 334
	if !ok {
		return false, errors.New(fmt.Sprintf(AUTH_LOGIN_ERROR, code, statusLine))
	}
	// TODO: Check it requests username
	// -----
	// Username
	b64Username := base64.StdEncoding.EncodeToString([]byte(mc.address))
	// Write usename
	if err = writeCRLFFlush(mc.smtpReadWriter, b64Username); err != nil {
		return false, fmt.Errorf(WRITE_FAILURE, err)
	}
	// Read 334 (password)
	line, err = readLine(mc.smtpReadWriter)
	if err != nil {
		return false, fmt.Errorf(READ_FAILURE, err)
	}
	// Handle resp
	ok, code, statusLine = handleSmtpResponse(line, 334)
	// Error out on a none 334
	if !ok {
		return false, errors.New(fmt.Sprintf(AUTH_LOGIN_ERROR, code, statusLine))
	}
	// Password
	b64Password := base64.StdEncoding.EncodeToString([]byte(mc.password))
	// Write password
	if err = writeCRLFFlush(mc.smtpReadWriter, b64Password); err != nil {
		return false, fmt.Errorf(WRITE_FAILURE, err)
	}
	// Check all is ok and authenticated (235)
	line, err = readLine(mc.smtpReadWriter)
	if err != nil {
		return false, fmt.Errorf(READ_FAILURE, err)
	}
	ok, code, statusLine = handleSmtpResponse(line, 235)
	return ok, err
}

// Opens a new smtp connection, if tls is wanted, optionally turn tls on
// attempts to login with basic auth from the initial credential provided
// when creating the client
func (mc *MailClient) OpenSMTPConnection(tls bool) (bool, error) {
	var err error
	if err = mc.connectSMTPInsecure(); err != nil {
		return false, err
	}
	if tls {
		if err = mc.upgradeSMTPConnectionTLS(); err != nil {
			return false, err
		}
	}
	ok, err := mc.loginBasic()
	return ok, err
}

// Attempts to close the smtp connection, if it isn't able,
// will return err else nil
func (mc *MailClient) CloseSMTPConnection() error {
	return mc.smtpConnection.Close()
}

// Sends mail to the given address securely
// returns whether or not the mail sent successfully and any errors that occured
func (mc *MailClient) SendNewMail(recipientAddress string, body string) (bool, error) {
	var err error
	// Set send address
	if err = writeCRLFFlush(mc.smtpReadWriter, fmt.Sprintf("MAIL FROM: %s", mc.address)); err != nil {
		return false, fmt.Errorf(WRITE_FAILURE, err)
	}
	line, err := readLine(mc.smtpReadWriter)
	if err != nil {
		return false, fmt.Errorf(READ_FAILURE, err)
	}
	// Need to send meta data to handler to explain what kind of '250' we expect, same for
	// other errors too.
	ok, code, statusLine := handleSmtpResponse(line, 250)
	if !ok {
		return false, errors.New(fmt.Sprintf(MAIL_FROM_ERROR, code, statusLine))
	}

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
	return true, err
}
