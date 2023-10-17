package mailer_lib

type Message struct {
	To          []string
	Title       string
	Template    *Template
	Attachments *[]Attachment `json:"attachments"`
}

type Template struct {
	PathBase string
	FileName string
	Data     interface{}
}

type Attachment struct {
	FileName string
	Data     []byte
}

type TemplatePath string

type ConfigMailer struct {
	AWSRegion string
	FromMail  string
}
