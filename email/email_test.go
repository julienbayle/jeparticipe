package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailTemplate(t *testing.T) {
	email := NewEmail("test@circuleo.fr", "Object", "A body")

	templateData := struct {
		URL string
	}{
		URL: "a_url",
	}
	assert.NotPanics(t, func() {
		email.AddBodyUsingTemplate("../templates/confirm.html", templateData)
	})
	assert.Contains(t, email.Body, "a_url")
	assert.Contains(t, email.Body, "confirmer")
}

func TestEmailTemplatePanics(t *testing.T) {
	email := NewEmail("test@circuleo.fr", "Object", "A body")

	templateData := struct {
		URL string
	}{
		URL: "a_url",
	}
	assert.Panics(t, func() {
		email.AddBodyUsingTemplate("../templates/donotexist.html", templateData)
	})
}
