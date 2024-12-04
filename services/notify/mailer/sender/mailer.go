package sender

import (
	"net/smtp"
)

type Mailer struct {
	From     string
	Password string
	SmtpHost string
	SmtpPort string
}


func (mailer *Mailer) SendMail(msg string, to string) error {
	
	subject := "Subject: HTML Email Test\n"
	
	contentType := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	
	auth := smtp.PlainAuth("", mailer.From, mailer.Password, mailer.SmtpHost)

	message := []byte(subject + contentType + msg)

	err := smtp.SendMail(mailer.SmtpHost+":"+mailer.SmtpPort, auth, mailer.From, []string{to}, message)
	
	if err != nil {
		return err
	}

	return nil
}