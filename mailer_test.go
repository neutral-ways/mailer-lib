package mailer_lib

import (
	"fmt"
	"testing"
)

func TestSendMail(t *testing.T) {
	config := ConfigMailer{
		AWSRegion:    "us-east-2",
		FromMail:     "no-reply@corezz.net",
		PathTemplate: "templates",
	}

	mailer := NewMailer(nil, config)
	msj := Message{
		To:    []string{"pdefilippis@gmail.com"},
		Title: "Test",
	}

	err := mailer.SendMail(msj)
	fmt.Print(err)
}
