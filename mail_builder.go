package main

import "fmt"

// Wraps an email message in a re-usable struct
// Often the data may change, but not the base fields such as TO/FROM/SUBJECT,
// in these events this struct can be reused with updated data
// hence the naming convention for UpdateMailBody(), using the verb 'update' rather than 'set'.
type MailBuilder struct {
	FROM    string
	TO      string
	SUBJECT string
	DATA    []byte
}

// Sets the FROM field of an email
func (mb *MailBuilder) SetFrom(fromName string, fromAddress string) *MailBuilder {
	mb.FROM = fmt.Sprintf("From: %s %s\r\n", fromName, fromAddress)
	return mb
}

// Sets the TO field of an email
func (mb *MailBuilder) SetTo(toAddress string) *MailBuilder {
	mb.TO = fmt.Sprintf("To: %s\r\n\r\n", toAddress)
	return mb
}

// Sets the SUBJECT field of an email
func (mb *MailBuilder) SetSubject(subject string) *MailBuilder {
	mb.SUBJECT = fmt.Sprintf("Subject: %s\r\n", subject)
	return mb
}

// Updates the mail body
func (mb *MailBuilder) UpdateMailBody(builder MailBodyBuilder) *MailBuilder {
	mb.DATA = builder.GetBody()
	return mb
}

// String update
func (mb *MailBuilder) UpdateMailBodyString(body string) *MailBuilder {
	mb.DATA = []byte(body)
	return mb
}

// String update
func (mb *MailBuilder) UpdateMailBodyBytes(body []byte) *MailBuilder {
	mb.DATA = body
	return mb
}

// Builds the body into a streamable email
func (mb *MailBuilder) Build() []byte {
	d := make([]byte, 0, 1000)
	d = append(d, []byte(mb.FROM)...)
	d = append(d, []byte(mb.SUBJECT)...)
	d = append(d, []byte(mb.TO)...)
	d = append(d, mb.DATA...)
	d = append(d, []byte("\r\n.\r\n")...)
	return d
}

// Provides basic utiliy for composing a message body
type MailBodyBuilder struct {
	DATA []byte
}

// Writes a line into the mail body followed by a carriage newline return (crlf)
func (mbb *MailBodyBuilder) WriteLine(data string) *MailBodyBuilder {
	mbb.DATA = append(mbb.DATA, []byte(data)...)
	mbb.DATA = append(mbb.DATA, 13, 10)
	return mbb
}

// Appends data to the current line
func (mbb *MailBodyBuilder) AppendToLine(data string) *MailBodyBuilder {
	mbb.DATA = append(mbb.DATA, []byte(data)...)
	return mbb
}

// Appends data to the current line
func (mbb *MailBodyBuilder) BreakToNewLine() *MailBodyBuilder {
	mbb.DATA = append(mbb.DATA, 13, 10)
	return mbb
}

// Gets the data stored internally
func (mbb *MailBodyBuilder) GetBody() []byte {
	return mbb.DATA
}
