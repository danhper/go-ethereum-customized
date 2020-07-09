package alerter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethdb"
)

var (
	DestinationsKey []byte = []byte("geth-alerter-destinations")
)

// EmailConfig holds the configuration necessary to send emails
type EmailConfig struct {
	FromEmail    string
	FromName     string
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
	db           ethdb.Database
}

// NewAlerter creates a new Alerter
func NewAlerter(config *Config, db ethdb.Database) *Alerter {
	return &Alerter{
		config:       config,
		destinations: make(map[string]Sender),
		db:           db,
	}
}

func (a *Alerter) loadDestinations() (destinations []string) {
	result, err := a.db.Get(DestinationsKey)
	if err != nil {
		return
	}
	json.Unmarshal(result, &destinations)
	return
}

func (a *Alerter) persistDestination(destination string) error {
	destinations := a.loadDestinations()
	destinations = append(destinations, destination)
	toWrite, err := json.Marshal(destinations)
	if err != nil {
		return err
	}
	return a.db.Put(DestinationsKey, toWrite)
}

// RegisterDestination registers a new destination to which to send
// the alert when one is triggered
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
	err := a.persistDestination(destination)
	return true, err
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
