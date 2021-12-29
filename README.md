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
	mailAddress := "<your email>@<your domain>"
	mailPwd := "<yourpassword>"
	client := mua.NewMailClient(mailAddress, mailPwd)
	client.UpgradeConnectionTLS()

	if succesfulLogin := client.LoginBasicSecure(); succesfulLogin {
		fmt.Println("login?: ", succesfulLogin)
	}

	mailBuilder := mua.MailBuilder{}
	mailBuilder.SetTo(mailAddress)
	mailBuilder.SetFrom("digletti", mailAddress)
	mailBuilder.SetSubject("testing this thing out")
	mailBuilder.UpdateMailBodyString("sending mail")

	client.SendNewMail(mailAddress, string(mailBuilder.Build()))
}
```
