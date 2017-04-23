package services

import (
	"crypto/rand"
	"io"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
)

const (
	SuperAdminLogin  = "superadmin"
	AdminLoginSuffix = "admin"
)

var (
	PasswordRune = []byte("ABCDEFGHIJKLPQRSTUVWXYZabcdefghijkpqrstuvwxyz23456789!@")
)

type SecurityToken struct {
	Token string `json:"token"`
}

// Checks that current user has admin priviledge
func hasAdminPriviledge(r *rest.Request) bool {
	user := r.Env["REMOTE_USER"]
	eventCode := getEventCodeFromRequest(r)

	if user == nil {
		return false
	}

	if user.(string) == SuperAdminLogin {
		return true
	}

	if eventCode != "" {
		return user.(string) == GetEventAdminLogin(eventCode)
	}

	return false
}

func GetEventAdminLogin(eventCode string) string {
	return eventCode + "-" + AdminLoginSuffix
}

func Authenticate(eventService *EventService, superAdminPassword string, userId string, password string) bool {
	if userId == SuperAdminLogin {
		return password == superAdminPassword
	}

	userIdParts := strings.Split(userId, "-")
	if len(userIdParts) == 2 && userIdParts[1] == AdminLoginSuffix {
		event := eventService.GetEvent(userIdParts[0])
		return event != nil && password == event.AdminPassword
	}

	return false
}

// NewPassword generates random passwords
// Inspired by "github.com/cmiceli/password-generator-go"
func NewPassword(length int) string {
	new_pword := make([]byte, length)
	random_data := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(PasswordRune))
	maxrb := byte(256 - (256 % len(PasswordRune)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, random_data); err != nil {
			panic(err)
		}
		for _, c := range random_data {
			if c >= maxrb {
				continue
			}
			new_pword[i] = PasswordRune[c%clen]
			i++
			if i == length {
				return string(new_pword)
			}
		}
	}
	panic("unable to generate password")
}
