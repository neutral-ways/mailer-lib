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
	"path/filepath"
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

	///Users/pdefilippis/Documents/proyectos/corezero/backend/fast_assesment/fast-assesment/internal/providers
	curdir, err := filepath.Abs("./templates")
	if err != nil {
		return "", err
	}

	var allPaths []string
	for _, t := range allTemplates {
		allPaths = append(allPaths, fmt.Sprintf("%s/%s", curdir+"/templates", t))
	}
	/*
		templates := template.Must(template.New("").Funcs(template.FuncMap{
			"safe": func(s string) template.HTML { return template.HTML(s) },
		}).ParseFiles(allPaths...))
	*/
	templates := template.Must(template.New("").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).Parse(mailer.get()))

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

func (mailer *Mailer) get() string {
	return "{{define \"page\"}}\n<!DOCTYPE html>\n<html lang=\"en\" xmlns=http://www.w3.org/1999/xhtml xmlns:v=\"urn:schemas-microsoft-com:vml\" xmlns:o=\"urn:schemas-microsoft-com:office:office\">\n<head>\n  <meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\"/>\n  {{safe \"<!--[if !mso]>\"}}\n  <meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\" />\n  {{safe \"<![endif]-->\"}}\n\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n  <style type=\"text/css\">\n      #outlook a { padding:0; }\n      body { margin:0;padding:0;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%; }\n      table, td { border-collapse:collapse; }\n      img { border:0;height:auto;line-height:100%; outline:none;text-decoration:none;-ms-interpolation-mode:bicubic; }\n      p { display:block;margin:0 0 6px; }\n  </style>\n\n  {{safe \"<!--[if mso]>\"}}\n  <noscript>\n    <xml>\n      <o:OfficeDocumentSettings>\n        <o:AllowPNG/>\n        <o:PixelsPerInch>96</o:PixelsPerInch>\n      </o:OfficeDocumentSettings>\n    </xml>\n  </noscript>\n  {{safe \"<![endif]-->\"}}\n\n  {{safe \"<!--[if lte mso 11]>\"}}\n  <style type=\"text/css\">\n    .mj-outlook-group-fix { width:100% !important; }\n  </style>\n  {{safe \"<![endif]-->\"}}\n\n  <style type=\"text/css\">\n    @media only screen and (min-width:480px) {\n      .mj-column-per-100 { width:100% !important; max-width: 100%; }\n      .mj-column-per-80 { width:80% !important; max-width: 80%; }\n      .mj-column-per-20 { width:20% !important; max-width: 20%; }\n    }\n    </style>\n\n    <style media=\"screen and (min-width:480px)\">\n      .moz-text-html .mj-column-per-100 { width:100% !important; max-width: 100%; }\n      .moz-text-html .mj-column-per-80 { width:80% !important; max-width: 80%; }\n      .moz-text-html .mj-column-per-20 { width:20% !important; max-width: 20%; }\n    </style>\n\n    <style type=\"text/css\">\n      @media only screen and (max-width:480px) {\n        table.mj-full-width-mobile { width: 100% !important; }\n        td.mj-full-width-mobile { width: auto !important; }\n    }\n  </style>\n\n</head>\n\n<body style=\"word-spacing:normal;background-color:#f5f5f5; padding: 20px 0; margin:0px auto;font-family:Arial, Helvetica Neue, Helvetica, sans-serif\">\n    {{template \"header\" -}}\n    {{template \"content\" . -}}\n    {{template \"footer\" -}}\n</body>\n</html>\n{{end}}\n{{define \"header\" -}}\n<div style=\"background-color:#f5f5f5; padding-top: 20px;\">\n    {{safe \"<!--[if mso | IE]>\"}}\n      <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" class=\"\" role=\"presentation\" style=\"width:600px;\" width=\"600\" bgcolor=\"FBFCFD\" >\n        <tr>\n          <td style=\"line-height:0px;font-size:0px;mso-line-height-rule:exactly;\">\n    {{safe \"<![endif]-->\"}}\n    <div style=\"background:#FBFCFD;background-color:#FBFCFD;margin:0px auto;max-width:600px;\">\n      <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" style=\"background:#FBFCFD;background-color:#FBFCFD;width:100%;\">\n        <tbody>\n          <tr>\n            <td style=\"direction:ltr;font-size:0px;padding:0 0 32px;text-align:center;\">\n              <img\n                  src=\"https://public-files-cz.s3.amazonaws.com/assets/corezero-email-header.jpg\"\n                  width=\"600\"\n                />\n            </td>\n          </tr>\n        </tbody>\n      </table>\n    </div>\n    {{safe \"<!--[if mso | IE]>\"}}\n          </td>\n        </tr>\n      </table>\n    {{safe \"<![endif]-->\"}}\n</div>\n{{end}}\n{{define \"footer\" -}}\n<div style=\"background-color:#f5f5f5; padding-bottom: 20px;\">\n    {{safe \"<!--[if mso | IE]>\"}}\n      <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" class=\"\" role=\"presentation\" style=\"width:600px;\" width=\"600\" bgcolor=\"FBFCFD\" >\n        <tr>\n          <td style=\"line-height:0px;font-size:0px;mso-line-height-rule:exactly;\">\n    {{safe \"<![endif]-->\"}}\n    <div style=\"background:#FBFCFD;background-color:#FBFCFD;margin:0px auto;max-width:600px;\">\n      <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" style=\"background:#FBFCFD;background-color:#FBFCFD;width:100%;\">\n        <tbody>\n          <tr>\n            <td style=\"direction:ltr;font-size:0px;padding:96px 32px 48px;text-align:left;\">\n              <p style=\"font-family:Arial, sans-serif;font-size:12px;line-height:1.2;color:#0D0D0D;\">Regards,</p>\n              <p style=\"font-family:Arial, sans-serif;font-size:12px;line-height:1.2;color:#0D0D0D;\"><strong>The CoreZero team</strong></p>\n            </td>\n          </tr>\n        </tbody>\n      </table>\n    </div>\n    {{safe \"<!--[if mso | IE]>\"}}\n          </td>\n        </tr>\n      </table>\n    {{safe \"<![endif]-->\"}}\n\n    {{safe \"<!--[if mso | IE]>\"}}\n      <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" class=\"\" role=\"presentation\" style=\"width:600px;\" width=\"600\" bgcolor=\"FBFCFD\" >\n        <tr>\n          <td style=\"line-height:0px;font-size:0px;mso-line-height-rule:exactly;\">\n    {{safe \"<![endif]-->\"}}\n    <div style=\"background:#0D0D0D;background-color:#0D0D0D;margin:0px auto;max-width:600px;\">\n      <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" style=\"background:#0D0D0D;background-color:#0D0D0D;width:100%;\">\n        <tbody>\n          <tr>\n            <td style=\"direction:ltr;font-size:0px;text-align:left;padding: 40px 32px 56px;\">\n              <div style=\"color: #FBFCFD; font-family:Arial, sans-serif; font-size: 16px; line-height: 1.2; text-transform: uppercase; max-width: 344px; width: 100%;\">\n                Unleashing the value of zero waste <br />\n                to tackle climate change\n              </div>\n            </td>\n          </tr>\n          <tr>\n            <td style=\"direction:ltr;font-size:0px;text-align:left;padding: 0 32px 20px;\">\n              <table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" style=\"background:#0D0D0D;background-color:#0D0D0D;width:100%;\">\n                <tr>\n                  <td width=\"25%\">\n                    <p style=\"font-family:Arial, sans-serif;font-size:10px;line-height:1.2;color:#FBFCFD;\">\n                      Get in touch <br />\n                      <a href=\"mailto:info@corezero.io\" style=\"color: #c5ff80;font-family:Arial, sans-serif;font-size: 10px;line-height: 1.2;\">info@corezero.io</a\n                      >\n                    </p>\n                  </td>\n                  <td width=\"40%\">\n                    <p style=\"font-family:Arial, sans-serif;font-size:10px;line-height:1.2;color:#FBFCFD;\">\n                      Stay tuned <br />\n                      <a href=\"https://www.linkedin.com/company/corezero\" style=\"color: #c5ff80;font-family:Arial, sans-serif;font-size: 10px;line-height: 1.2;\">LinkedIn&#8599;</a\n                      >\n                    </p>\n                  </td>\n                  <td>\n                    <p style=\"font-family:Arial, sans-serif;font-size:10px;line-height:1.2;color:#FBFCFD;\">&copy; 2023 CoreZero. All rights reserved.</p>\n                    <p style=\"font-family:Arial, sans-serif;font-size:8px;line-height:1.2;color:#FBFCFD;\">\n                      <span>3350 Virginia Street 2nd Floor, Office 222,</span>\n                      <span>Coconut Grove, FL 33133 </span>\n                    </p>\n                  </td>\n                </tr>\n              </table>\n            </td>\n          </tr>\n        </tbody>\n      </table>\n    </div>\n    {{safe \"<!--[if mso | IE]>\"}}\n          </td>\n        </tr>\n      </table>\n    {{safe \"<![endif]-->\"}}\n</div>\n{{end}}"
}
