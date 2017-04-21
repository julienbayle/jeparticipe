package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailTemplate(t *testing.T) {
	email := NewEmail("jbayle@gmail.com", "Object", "A body")

	templateData := struct {
		URL string
	}{
		URL: "a_url",
	}
	assert.NoError(t, email.AddBodyUsingTemplate("../templates/confirm.html", templateData))
	assert.Contains(t, email.Body, "a_url")
	assert.Contains(t, email.Body, "confirmer")
}
