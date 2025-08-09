package contactsedit

import (
	"strings"

	"github.com/a-h/templ/examples/crud/db"
)

func NewModel() Model {
	return Model{
		Initial: true,
	}
}

func ModelFromContact(contact db.Contact) (m Model) {
	return Model{
		Initial: true,
		Name:    contact.Name,
		Email:   contact.Email,
	}
}

type Model struct {
	Initial          bool
	SubmitButtonText string

	Name  string
	Email string
	Error string
}

func (m *Model) ValidateName() (msgs []string) {
	if m.Initial {
		return
	}
	if m.Name == "" {
		msgs = append(msgs, "Name is required")
	}
	return msgs
}

func (m *Model) NameHasError() bool {
	return len(m.ValidateName()) > 0
}

func (m *Model) ValidateEmail() (msgs []string) {
	if m.Initial {
		return
	}
	if m.Email == "" {
		return append(msgs, "Email is required")
	}
	if !strings.Contains(m.Email, "@") {
		msgs = append(msgs, "Email is invalid")
	}
	return msgs
}

func (m *Model) EmailHasError() bool {
	return len(m.ValidateEmail()) > 0
}

func (m *Model) Validate() (msgs []string) {
	if m.Initial {
		return
	}
	msgs = append(msgs, m.ValidateName()...)
	msgs = append(msgs, m.ValidateEmail()...)
	return msgs
}
