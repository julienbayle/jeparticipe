package services

import (
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
)

const (
	SuperAdminLogin  = "superadmin"
	AdminLoginSuffix = "admin"
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
