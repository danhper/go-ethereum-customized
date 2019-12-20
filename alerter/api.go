package alerter


// PublicAlerterAPI exposes the functionality of Alerter to the RPC client
type PublicAlerterAPI struct {
	alerter *Alerter
}


// NewPublicAlerterAPI create a new PublicAlerterAPI.
func NewPublicAlerterAPI(alerter *Alerter) *PublicAlerterAPI {
	return &PublicAlerterAPI{
		alerter: alerter,
	}
}


// RegisterDestination delegates to Alerter.RegisterDestination
// Destination should have the following format:
// transport:endpoint
// where transport could be 'http' or 'smtp' (note that https urls work with 'http' endpoint)
// for example
// http:https://example.com/alert
// smtp:alert@example.com
// SMTP configuration must be set through command line options for
// the STMP transport to work
func (api *PublicAlerterAPI) RegisterDestination(destination string) (bool, error) {
	return api.alerter.RegisterDestination(destination)
}

// ListDestinations delegates to Alerter.ListDestinations
func (api *PublicAlerterAPI) ListDestinations() ([]string, error) {
	return api.alerter.ListDestinations()
}

// SendTestAlert delegates to Alerter.SendAlert
func (api *PublicAlerterAPI) SendTestAlert(subject string, message string) error {
	return api.alerter.SendAlert(subject, message)
}
