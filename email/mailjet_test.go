package email

/**
import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test only pass if mailjet API credentials are in the path
func TestSendEmail(t *testing.T) {
	email := NewEmail("jbayle@gmail.com", "Object", "A body")

	emailRelay := &EmailRelay{
		Send: SendWithMailjet,
	}

	assert.NoError(t, emailRelay.Send(email))
}*/
