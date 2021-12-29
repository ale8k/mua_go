package main

import (
	"fmt"

	"github.com/ale8k/mua_go/mua"
)

// Just to test package locally & develop
func main() {
	mailAddress := "bob@bob.com"
	mailPwd := ""
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
