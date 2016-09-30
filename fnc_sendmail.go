package main

import (
	"strconv"

	"github.com/go-gomail/gomail"
)

var opts_sendmail map[string]string

func initSendMail() {
	opts_sendmail = make(map[string]string, 1)
	opts_sendmail["from"] = "uzkhinta@gmail.com"
	opts_sendmail["host"] = "smtp.gmail.com"
	opts_sendmail["port"] = "25"

	opts_sendmail["host"] = "smtp.gmail.com"
	opts_sendmail["port"] = "465"

	opts_sendmail["login"] = "uzkhinta"
	opts_sendmail["pass"] = "AsPeefW2m42i03yqVB9f123"
	opts_sendmail["bodytype"] = "text/html"
}

func SendMail(mailto, subject, msg string) {
	m := gomail.NewMessage()
	m.SetHeader("From", opts_sendmail["from"])
	m.SetHeader("To", mailto)
	m.SetHeader("Subject", subject)
	m.SetBody(opts_sendmail["bodytype"], msg)

	smtpport, _ := strconv.Atoi(opts_sendmail["port"])

	d := gomail.NewPlainDialer(
		//"smtp.gmail.com", 25, "", "")
		opts_sendmail["host"], smtpport, opts_sendmail["login"], opts_sendmail["pass"])

	err := d.DialAndSend(m)

	LogPrintErrAndExit("ОШИБКА отправки сообщения: ", err)
	LogPrint("письмо успешно отправлено")
}
