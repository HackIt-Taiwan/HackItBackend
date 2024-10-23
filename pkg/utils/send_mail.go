package utils

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	mail "github.com/wneessen/go-mail"
)

var (
	emailClient *mail.Client
	clientOnce  sync.Once
)

func InitEmailClient() error {
	var err error
	clientOnce.Do(func() {
		num, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
		if err != nil {
			err = fmt.Errorf("failed to convert mail port to int: %w", err)
			return
		}

		emailClient, err = mail.NewClient(os.Getenv("MAIL_SERVER"),
			mail.WithPort(num),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(os.Getenv("MAIL_USERNAME")),
			mail.WithPassword(os.Getenv("MAIL_PASSWORD")),
			mail.WithSSL(),                 // 啟用 SSL 連接
			mail.WithTLSPolicy(mail.NoTLS), // 關閉 TLS
		)
		if err != nil {
			err = fmt.Errorf("failed to create mail client: %w", err)
		}
	})

	return err
}

func SendEmail(to string, subject string, body string) error {
	if emailClient == nil {
		if err := InitEmailClient(); err != nil {
			return err
		}
	}

	// 建立郵件訊息
	msg := mail.NewMsg()
	msg.From(os.Getenv("MAIL_DEFAULT_SENDER"))
	msg.To(to)
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	// 發送郵件
	if err := emailClient.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
