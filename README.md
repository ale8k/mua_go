# mua_go
[![Go Reference](https://pkg.go.dev/badge/github.com/ale8k/mua_go.svg)](https://pkg.go.dev/github.com/ale8k/mua_go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ale8k/mua_go)](https://goreportcard.com/report/github.com/ale8k/mua_go)
[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/ale8k/mua_go/releases/latest)
![Tests](https://github.com/ale8k/mua_go/actions/workflows/tests.yml/badge.svg)

## Summary
A basic SMTP client to send and receive emails for a given address(s).

Sending an email example:
```go
func main() {
	mailAddress := ""
	mailPwd := ""

	client := mua.NewMailClient(mailAddress, mailPwd, "smtp-mail.outlook.com:587")
	client.OpenSMTPConnection(true)

	mailBuilder := mua.MailBuilder{}
	mailBuilder.SetTo(mailAddress)
	mailBuilder.SetFrom("digletti", mailAddress)
	mailBuilder.SetSubject("test 3")
	mailBuilder.UpdateMailBodyString("sending mail")

	client.SendNewMail(mailAddress, string(mailBuilder.Build()))

	mailBuilder.SetSubject("test 4")
	client.SendNewMail(mailAddress, string(mailBuilder.Build())) // isnt sending, huh ?
	err := client.CloseSMTPConnection()
	fmt.Println(err)
}
```
