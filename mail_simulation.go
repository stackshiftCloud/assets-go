package stackshift

type MailMode string

const (
	MailLive MailMode = "live"
	MailTest MailMode = "test"
)

type MailSimulationOptions struct {
	Scenario string `json:"scenario"`
}
