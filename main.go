package main

import (
	"fmt"

	"github.com/ale8k/mua_go/mua"
)

func main() {
	mailAddress := ""
	mailPwd := ""

	client := mua.NewMailClient(mailAddress, mailPwd, "smtp-mail.outlook.com:587")
	client.OpenSMTPConnection(true)

	mailBuilder := mua.MailBuilder{}
	mailBuilder.SetTo(mailAddress)
	mailBuilder.SetFrom("digletti", mailAddress)
	mailBuilder.SetSubject("test 1")
	mailBuilder.UpdateMailBodyString("sending mail")

	client.SendNewMail(mailAddress, string(mailBuilder.Build()))

	mailBuilder.SetSubject("test 2")
	client.SendNewMail(mailAddress, string(mailBuilder.Build())) // isnt sending, huh ?
	err := client.CloseSMTPConnection()
	fmt.Println(err)
}
