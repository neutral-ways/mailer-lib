package mailer_lib

import (
	"github.com/magiconair/properties/assert"
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
	assert.Equal(t, err, nil)
}
