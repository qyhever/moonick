package postal

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"moonick/internal/config"
	"net/smtp"
	"strings"
)

// SendMail 发送邮件
// to: 收件人邮箱
// subject: 邮件标题
// body: 邮件内容(支持HTML格式)
func SendMail(to, subject, body string) error {
	cfg := config.GetConfig()
	fromName := cfg.Postal.FromName
	fromEmail := cfg.Postal.FromEmail
	smtpServer := cfg.Postal.SmtpServer
	smtpPort := cfg.Postal.SmtpPort
	fromPass := cfg.Postal.FromPass
	// 构造邮件头
	header := make(map[string]string)
	header["From"] = fmt.Sprintf(`"%s" <%s>`, fromName, fromEmail)
	header["To"] = to
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
	header["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range header {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n" + body)

	// 建立 TLS 连接（465端口专用）
	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, &tls.Config{
		InsecureSkipVerify: true, // 忽略证书验证（VPS自签证书常用）
		ServerName:         smtpServer,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return err
	}
	defer client.Quit()

	// 认证
	if err := client.Auth(smtp.PlainAuth("", fromEmail, fromPass, smtpServer)); err != nil {
		return err
	}
	// 设置发件/收件
	if err := client.Mail(fromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	// 发送数据
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg.String()))
	if err != nil {
		return err
	}
	w.Close()

	return nil
}
