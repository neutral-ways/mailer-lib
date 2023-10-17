package mailer_lib

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"go.uber.org/zap"
	"go/build"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"reflect"
	"runtime"
)

const (
	footer       = "footer.tmpl"
	header       = "header.tmpl"
	contentEmpty = "content_empty.tmpl"
	page         = "page.tmpl"
)

var basicTemplates = []string{page, header, footer}

type Mailer struct {
	Log    *zap.Logger
	Config ConfigMailer
}

func NewMailer(log *zap.Logger, config ConfigMailer) *Mailer {
	return &Mailer{
		Log:    log,
		Config: config,
	}
}

func (mailer *Mailer) SendMail(msg Message) error {
	fromMail := fmt.Sprintf("CoreZero <%s>", mailer.Config.FromMail)
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
	allPaths, err := mailer.getAllPaths(tmpl)
	mailer.Log.Info("all paths", zap.Any("paths", allPaths))
	mailer.Log.Info("error all paths", zap.Error(err))
	if err != nil {
		return "", err
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
		mailer.Log.Info("error execute template", zap.Error(err))
		return "", err
	}

	return string(processed.Bytes()), nil
}

func (mailer *Mailer) getBasicTemplates(tmpl *Template) ([]string, error) {
	if tmpl == nil {
		basicTemplates = append(basicTemplates, contentEmpty)
	}

	/*
		pathBase, err := mailer.getCurrentPath()
		mailer.Log.Info("get current path", zap.String("path base", pathBase))
		if err != nil {
			mailer.Log.Error("error get current path", zap.Error(err))
			return nil, err
		}
	*/
	pathBase := tmpl.Path
	var allPaths []string
	for _, t := range basicTemplates {
		mailer.Log.Info("path created", zap.String("path", filepath.Join(pathBase, t)))
		allPaths = append(allPaths, filepath.Join(pathBase, t))
	}

	return allPaths, nil
}

func (mailer *Mailer) getCurrentPath() (string, error) {
	_, b, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(b), "../..")
	mailer.Log.Info("root", zap.String("root path", root))

	var t Template
	packagePath := reflect.TypeOf(t).PkgPath()
	detailPackage, err := build.Default.Import(packagePath, ".", build.FindOnly)
	//aca esta el error
	if err != nil {
		return "", err
	}

	return detailPackage.Dir, nil
}

func (mailer *Mailer) getAllPaths(tmpl *Template) ([]string, error) {
	allTemplates, err := mailer.getBasicTemplates(tmpl)
	if err != nil {
		return nil, err
	}

	if tmpl != nil {
		allTemplates = append(allTemplates, tmpl.Path)
	}

	return allTemplates, nil
}
