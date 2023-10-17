package mailer_lib

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"log"
	"os"
)

const (
	footer       = "footer.tmpl"
	header       = "header.tmpl"
	contentEmpty = "content_empty.tmpl"
	page         = "page.tmpl"
)

var basicTemplates = []string{page, header, footer}

type Mailer struct {
	log    *zap.Logger
	config ConfigMailer
}

func NewMailer(log *zap.Logger, config ConfigMailer) *Mailer {
	return &Mailer{
		log:    log,
		config: config,
	}
}

func (mailer *Mailer) SendMail(msg Message) error {
	fromMail := fmt.Sprintf("CoreZero <%s>", mailer.config.FromMail)
	messageBody, err := mailer.create(msg.Template)
	if err != nil {
		log.Println("Error occurred while creating email body", err)
		return fmt.Errorf("error occurred while creating email body %s", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-2"))
	if err != nil {
		log.Println("Error occurred while creating aws session", err)
		return fmt.Errorf("error occurred while creating aws session %s", err)
	}

	svc := ses.NewFromConfig(cfg)
	message := gomail.NewMessage()
	message.SetHeader("From", fromMail)
	message.SetHeader("To", msg.To...)
	message.SetHeader("Subject", msg.Title)
	message.SetBody("text/html", messageBody)

	// check just in case
	if msg.Attachments != nil && len(*msg.Attachments) > 0 {
		for _, a := range *msg.Attachments {
			message.Attach(a.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(a.Data)
				return err
			}))
		}
	}

	var emailRaw bytes.Buffer
	_, err = message.WriteTo(&emailRaw)
	if err != nil {
		log.Println("error occurred while writing email", zap.Error(err))
		return fmt.Errorf("error occurred while writing email to buffer %s", err)
	}

	_, err = svc.SendRawEmail(context.Background(), &ses.SendRawEmailInput{
		RawMessage: &types.RawMessage{
			Data: emailRaw.Bytes(),
		},
	})
	if err != nil {
		log.Println("error occurred while sending email", zap.Error(err))
		return fmt.Errorf("error occurred while sending email %s", err)
	}

	log.Println("Email sent successfully to: ", msg.To)
	return nil
}

func (mailer *Mailer) create(tmpl *Template) (string, error) {
	allTemplates := mailer.getBasicTemplates(tmpl)

	curdir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var allPaths []string
	for _, t := range allTemplates {
		allPaths = append(allPaths, fmt.Sprintf("%s/%s", curdir+"/templates", t))
	}

	templates := template.Must(template.New("").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).ParseFiles(allPaths...))
	var processed bytes.Buffer

	var data interface{} = nil
	if tmpl != nil {
		data = tmpl.Data
	}

	err = templates.ExecuteTemplate(&processed, "page", data)
	if err != nil {
		return "", err
	}

	return string(processed.Bytes()), nil
}

func (mailer *Mailer) getBasicTemplates(tmpl *Template) []string {
	if tmpl == nil {
		return append(basicTemplates, contentEmpty)
	}

	return append(basicTemplates, tmpl.Path)
}