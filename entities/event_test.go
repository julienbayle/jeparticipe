package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	// Everything should be fine
	event, err := NewPendingConfirmationEvent("code_test", "ip", "email@email.com")
	assert.Nil(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "code_test", event.Code)
	assert.Equal(t, "ip", event.CreatedBy)
	assert.Equal(t, "email@email.com", event.UserEmail)
	assert.True(t, event.CreatedAt.Before(time.Now()))
	assert.False(t, event.EmailConfirmed)
	assert.Len(t, event.AdminPassword, 8)

	// Invalid code
	event, err = NewPendingConfirmationEvent("c", "ip", "email@email.com")
	assert.NotNil(t, err)
	assert.Nil(t, event)

	// Invalid email
	event, err = NewPendingConfirmationEvent("code_test", "ip", "email")
	assert.NotNil(t, err)
	assert.Nil(t, event)
}

func TestConfirmationCode(t *testing.T) {
	event, err := NewPendingConfirmationEvent("code_test", "ip", "email@email.com")
	assert.Nil(t, err)
	assert.Len(t, event.ConfirmCode("secret"), 64)

	event2, _ := NewPendingConfirmationEvent("code_test", "ip", "email@email.com")
	assert.Equal(t, event.ConfirmCode("secret"), event2.ConfirmCode("secret"))

	event3, _ := NewPendingConfirmationEvent("code_test", "ip", "email2@email.com")
	assert.NotEqual(t, event.ConfirmCode("secret"), event3.ConfirmCode("secret"))

}
