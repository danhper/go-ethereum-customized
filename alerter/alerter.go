package alerter

import (
	"fmt"
	"strings"
)

// EmailConfig holds the configuration necessary to send emails
type EmailConfig struct {
	FromEmail         string
	FromName         string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

// Config holds the necessary configuration to send the alerts
type Config struct {
	Email EmailConfig
}

// Alerter contains all the logic to register and send alerts
type Alerter struct {
	config       *Config
	destinations map[string]Sender
}

// NewAlerter creates a new Alerter
func NewAlerter(config *Config) *Alerter {
	return &Alerter{
		config:       config,
		destinations: make(map[string]Sender),
	}
}

// RegisterDestination registers a new destination to which to send
// the alert when one is triggered
// TODO: use persistent storage
func (a *Alerter) RegisterDestination(destination string) (bool, error) {
	if _, ok := a.destinations[destination]; ok {
		return false, nil
	}
	splitted := strings.SplitN(destination, ":", 2)
	transport := splitted[0]
	endpoint := splitted[1]
	senderFactory, ok := senders[transport]
	if !ok {
		return false, fmt.Errorf("unknown transport type %s", transport)
	}
	a.destinations[destination] = senderFactory(endpoint, a.config)
	return true, nil
}

// ListDestinations returns the list of registered destination
func (a *Alerter) ListDestinations() (destinations []string, err error) {
	for destination := range a.destinations {
		destinations = append(destinations, destination)
	}
	return
}

// SendAlert send alerts to all the registered destinations
func (a *Alerter) SendAlert(subject string, message string) error {
	var errors []string
	for _, sender := range a.destinations {
		err := sender.Send(subject, message)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf("some destination failed: %s", strings.Join(errors, "; "))
}
