package pochta

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
)

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}

func SendMail(smtpserver string, auth smtp.Auth, from mail.Address, to mail.Address, subject string, body string) error {
	// setup a map for the headers
	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = subject

	// setup the message
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// create the smtp connection
	c, err := smtp.Dial(smtpserver)
	if err != nil {
		return errors.New("smtp.Dial: " + err.Error())
	}

	// set some TLS options, so we can make sure a non-verified cert won't stop us sending
	host, _, _ := net.SplitHostPort(smtpserver)
	tlc := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err = c.StartTLS(tlc); err != nil {
		return errors.New("client.StartTLS: " + err.Error())
	}

	// auth stuff
	if err = c.Auth(auth); err != nil {
		return errors.New("client.Auth: " + err.Error())
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return errors.New("client.Mail(" + from.Address + "): " + err.Error())
	}
	if err = c.Rcpt(to.Address); err != nil {
		return errors.New("client.Rcpt(" + to.Address + "): " + err.Error())
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return errors.New("client.Data: " + err.Error())
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return errors.New("writeCloser.Write: " + err.Error())
	}
	err = w.Close()
	if err != nil {
		return errors.New("writeCloser.Close: " + err.Error())
	}
	c.Quit()
	return nil
}
