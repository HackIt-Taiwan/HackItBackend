package utils

import (
	"fmt"
	"os"
	"strconv"

	mail "github.com/wneessen/go-mail"
)

func SendEmail(to string, subject string, body string) error {
	// 設定 SMTP 伺服器
	num, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		return fmt.Errorf("failed to convert mail port to int: %w", err)
	}

	client, err := mail.NewClient(os.Getenv("MAIL_SERVER"),
		mail.WithPort(num),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("MAIL_USERNAME")),
		mail.WithPassword(os.Getenv("MAIL_PASSWORD")),
		mail.WithSSL(),                 // 啟用 SSL 連接
		mail.WithTLSPolicy(mail.NoTLS), // 關閉 TLS
	)

	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	// 建立郵件訊息
	msg := mail.NewMsg()
	msg.From(os.Getenv("MAIL_DEFAULT_SENDER"))
	msg.To(to)
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	// 發送郵件
	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
