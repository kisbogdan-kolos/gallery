package email

import (
	"fmt"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/kisbogdan-kolos/gallery/backend/helper"
)

var auth smtp.Auth
var server string
var mailFrom string

func Init() error {
	server = helper.EnvGet("SMTP_SERVER", "email-smtp.eu-west-1.amazonaws.com:587")
	mailFrom = helper.EnvGet("SMTP_MAIL_FROM", "")

	parts := strings.Split(server, ":")
	if len(parts) != 2 {
		return fmt.Errorf("SMTP server format incorrect: must include a port")
	}

	username := helper.EnvGet("SMTP_USER", "")
	password := helper.EnvGet("SMTP_PASS", "")

	auth = smtp.PlainAuth("", username, password, parts[0])

	return nil
}

func Send(recipient string, subject string, message string) error {
	if strings.ContainsAny(recipient, "\r\n\t") {
		return fmt.Errorf("recipient format incorrect: must not include newline or tabs")
	}

	if strings.ContainsAny(subject, "\r\n\t") {
		return fmt.Errorf("subject format incorrect: must not include newline or tabs")
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", mailFrom, recipient, subject, message)

	return smtp.SendMail(server, auth, stripMail(mailFrom), []string{stripMail(recipient)}, []byte(msg))
}

func stripMail(emailAddress string) string {
	re := regexp.MustCompile("^.*?<(.*?)>$")

	if re.Match([]byte(emailAddress)) {
		return re.FindStringSubmatch(emailAddress)[1]
	}

	return emailAddress
}
