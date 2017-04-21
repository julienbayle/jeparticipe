package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"time"

	"github.com/cmiceli/password-generator-go"
)

const (
	EventsBucketName        = "events"
	EmailRegExp      string = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
)

type Event struct {
	Code           string
	CreatedAt      time.Time
	CreatedBy      string
	UserEmail      string
	EmailConfirmed bool
	AdminPassword  string
}

// Creates a new pending confirmation event
func NewPendingConfirmationEvent(code string, ip string, userEmail string) (*Event, error) {
	eventCodeValidator, _ := regexp.Compile("[-A-Za-z0-9]{2,50}")
	emailValidator, _ := regexp.Compile(EmailRegExp)

	if !eventCodeValidator.MatchString(code) {
		return nil, errors.New("Invalid code")
	}

	if !emailValidator.MatchString(userEmail) {
		return nil, errors.New("Invalid email")
	}

	return &Event{
		Code:           code,
		CreatedAt:      time.Now(),
		CreatedBy:      ip,
		UserEmail:      userEmail,
		EmailConfirmed: false,
		AdminPassword:  pwordgen.NewPassword(8),
	}, nil
}

func (event *Event) ConfirmCode(secret string) string {
	h := sha256.New()
	h.Write([]byte(event.Code + event.UserEmail + secret))
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes[:])
}
