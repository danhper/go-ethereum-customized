package alerter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"strings"
)

type senderFactory func(endpoint string, cfg *Config) Sender

var senders = map[string]senderFactory{
	"http": NewHTTPSender,
	"smtp": NewSMTPSender,
}

// Sender is a common interface for multiple alert
// backends such as SMTP or HTTP
type Sender interface {
	Send(subject string, message string) error
}


// HTTPSender is the backend to send notifications through HTTP
// TODO: allow to customize headers, format and whatnot
type HTTPSender struct {
	url    string
	client *http.Client
}

// NewHTTPSender returns a new sender for the given url
func NewHTTPSender(url string, _cfg *Config) Sender {
	return &HTTPSender{
		url:    url,
		client: &http.Client{},
	}
}

// Send executes an HTTP request to the given endpoint
// TODO: this is currently made to work with slack incoming hook
// it should be made more customizable format wise
func (sender *HTTPSender) Send(subject string, message string) error {
	payload := map[string]string{"text": subject + "\n" + message}
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", sender.url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := sender.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

// SMTPSender is the backend to send notifications through SMTP
type SMTPSender struct {
	email string
	cfg   *EmailConfig
}

// NewSMTPSender returns a new sender for the given url
func NewSMTPSender(email string, cfg *Config) Sender {
	return &SMTPSender{
		email: email,
		cfg:   &cfg.Email,
	}
}

// Send sends an email via the configured smtp settings
func (sender *SMTPSender) Send(subject string, message string) error {
	headers := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", sender.cfg.FromName, sender.cfg.FromEmail),
		fmt.Sprintf("To: %s", sender.email),
		fmt.Sprintf("Subject: %s", subject),
	}, "\n")
	payload := []byte(headers + "\n\n" + message)

	smtpEndpoint := fmt.Sprintf("%s:%d", sender.cfg.SMTPHost, sender.cfg.SMTPPort)
	// TODO: allow to customize auth
	auth := smtp.PlainAuth("", sender.cfg.SMTPUser, sender.cfg.SMTPPassword, sender.cfg.SMTPHost)
	return smtp.SendMail(smtpEndpoint, auth, sender.cfg.FromEmail, []string{sender.email}, payload)
}
