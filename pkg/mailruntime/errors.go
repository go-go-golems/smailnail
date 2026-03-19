package mailruntime

import "fmt"

// MailError is a structured error returned by IMAP or Sieve operations.
type MailError struct {
	Name    string                 `json:"name"`
	Message string                 `json:"message"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	Source  string                 `json:"source"`
}

func (e *MailError) Error() string {
	return fmt.Sprintf("[%s/%s] %s", e.Source, e.Name, e.Message)
}

func NewMailError(name, message, source string) *MailError {
	return &MailError{Name: name, Message: message, Source: source}
}

func NewSieveError(name, message string) *MailError {
	return &MailError{Name: name, Message: message, Source: "sieve"}
}

func NewIMAPError(name, message string) *MailError {
	return &MailError{Name: name, Message: message, Source: "imap"}
}
