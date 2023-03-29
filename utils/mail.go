package utils

import (
	"ggslayer/utils/config"
	"gopkg.in/gomail.v2"
	"strings"
)

type mail struct {
	MailHost string
	MailPort int
	MailUser string // 发件人
	MailPass string // 发件人密码
	MailTo   string // 收件人 多个用,分割
	Subject  string // 邮件主题
	Body     string // 邮件内容
}

func NewMail() *mail {
	mailHost := config.GetString("mail.Host")
	mailPort := config.GetInt("mail.Port")
	mailUser := config.GetString("mail.User")
	mailPass := config.GetString("mail.Pass")
	return &mail{
		MailHost: mailHost,
		MailPort: mailPort,
		MailUser: mailUser,
		MailPass: mailPass,
	}
}

func (m *mail) SetMailInfo(mailTo, subject, body string) *mail {
	m.MailTo = mailTo
	m.Subject = subject
	m.Body = body
	return m
}

func (m *mail) Send() error {
	msg := gomail.NewMessage()
	//设置发件人
	msg.SetHeader("From", m.MailUser)
	//设置发送给多个用户
	mailArrTo := strings.Split(m.MailTo, ",")
	msg.SetHeader("To", mailArrTo...)
	//设置邮件主题
	msg.SetHeader("Subject", m.Subject)
	//设置邮件正文
	msg.SetBody("text/html", m.Body)
	d := gomail.NewDialer(m.MailHost, m.MailPort, m.MailUser, m.MailPass)
	return d.DialAndSend(msg)
}
