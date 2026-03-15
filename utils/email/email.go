package email

import (
	"fmt"
	"net/smtp"
	"voyagear/config"
)

var smtpCfg config.SMTPConfig

func Init (cfg config.SMTPConfig) {
	smtpCfg = cfg
}

func SendOTP(to string, otp string) error {

	from := smtpCfg.Username

	fromHeader := smtpCfg.From

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"subject: Your OTP for Voyagear\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"+
			"Hello!\n\nYour OTP is: %s\nIt will expire in 5 minutes.\n\nThanks,\nVestra Ecommerce Team",
			fromHeader, to, otp,
	)

	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)

	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
		return err
	}

	return nil
}