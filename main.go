package main

import (
	"github.com/ale8k/mua_go/mua"
)

// Just to test package locally & develop
func main() {
	mailAddress := ""
	mailPwd := ""
	client := mua.NewMailClient(mailAddress, mailPwd)

	mailBuilder := mua.MailBuilder{}
	mailBuilder.SetTo(mailAddress)
	mailBuilder.SetFrom("digletti", mailAddress)
	mailBuilder.SetSubject("testing this thing out")
	mailBuilder.UpdateMailBodyString("sending mail")

	client.SendNewMail(mailAddress, string(mailBuilder.Build()))
}
