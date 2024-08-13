package interactionsapi

const (
	errTypePrefix = "https://ffxiv.c032.dev/discord#error/"

	ErrTypeInternalServerError = errTypePrefix + "internal-server-error"
	ErrTypeUnknownCommand      = errTypePrefix + "unknown-command"
)

// ErrorResponse is an object as defined by RFC 7807.
type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}
