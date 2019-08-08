package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
)

type EmailerType struct {
	ServerName string
	Username   string
	Password   string
}

var Emailer EmailerType

// 初始化
// servername 如：smtp.exmail.qq.com:465
func InitEmailer(servername string, username string, password string) {
	Emailer.Password = password
	Emailer.Username = username
	Emailer.ServerName = servername
}

// send email over SSL
func (emailer EmailerType) SendToMail(toAddr string, subject string, body string) (err error) {
	host, _, _ := net.SplitHostPort(emailer.ServerName)
	// get SSL connection
	conn, err := dial(emailer.ServerName)
	if err != nil {
		return
	}
	// create new SMTP client
	smtpClient, err := smtp.NewClient(conn, host)
	if err != nil {
		return
	}
	// Set up authentication information.
	auth := smtp.PlainAuth("", emailer.Username, emailer.Password, host)
	// auth the smtp client
	err = smtpClient.Auth(auth)
	if err != nil {
		return
	}
	// set To && From address, note that from address must be same as authorization user.
	from := mail.Address{Name: "", Address: emailer.Username}
	to := mail.Address{Name: "", Address: toAddr}
	err = smtpClient.Mail(from.Address)
	if err != nil {
		return
	}
	err = smtpClient.Rcpt(to.Address)
	if err != nil {
		return
	}
	// Get the writer from SMTP client
	writer, err := smtpClient.Data()
	if err != nil {
		return
	}
	// compose message body
	message := composeMsg(from.String(), to.String(), subject, body)
	// write message to recp
	_, err = writer.Write([]byte(message))
	if err != nil {
		return
	}
	// close the writer
	err = writer.Close()
	if err != nil {
		return
	}
	// Quit sends the QUIT command and closes the connection to the server.
	_ = smtpClient.Quit()
	return nil
}

// dial using TLS/SSL
func dial(addr string) (*tls.Conn, error) {
	return tls.Dial("tcp", addr, nil)
}

// compose message according to "from, to, subject, body"
func composeMsg(from string, to string, subject string, body string) (message string) {
	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	// Setup message
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	return
}
