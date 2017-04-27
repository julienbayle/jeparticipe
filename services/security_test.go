package services

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/julienbayle/jeparticipe/entities"
	"github.com/stretchr/testify/assert"

	"net/http"
	"os"
	"testing"
)

func NewRequest() *rest.Request {
	origReq, _ := http.NewRequest("", "", nil)
	r := &rest.Request{
		origReq,
		make(map[string]string, 1),
		make(map[string]interface{}, 1),
	}
	return r
}

func TestAdminPriviledge(t *testing.T) {
	r := NewRequest()
	r.PathParams["event"] = "testevent"
	r.Env["REMOTE_USER"] = "testevent-admin"
	assert.True(t, hasAdminPriviledge(r))
	assert.False(t, hasSuperAdminPriviledge(r))
}

func TestSuperadminPriviledge(t *testing.T) {
	r := NewRequest()
	r.PathParams["event"] = "testevent"
	r.Env["REMOTE_USER"] = "superadmin"
	assert.True(t, hasAdminPriviledge(r))

	r = NewRequest()
	r.Env["REMOTE_USER"] = "superadmin"
	assert.True(t, hasAdminPriviledge(r))
	assert.True(t, hasSuperAdminPriviledge(r))
}

func TestNoPriviledge(t *testing.T) {
	r := NewRequest()
	r.PathParams["event"] = "testevent"
	r.Env["REMOTE_USER"] = "auser"
	assert.False(t, hasAdminPriviledge(r))

	r = NewRequest()
	r.PathParams["event"] = "testevent"
	r.Env["REMOTE_USER"] = "otherevent"
	assert.False(t, hasAdminPriviledge(r))

	r = NewRequest()
	assert.False(t, hasAdminPriviledge(r))

	r = NewRequest()
	r.Env["REMOTE_USER"] = "admin"
	assert.False(t, hasAdminPriviledge(r))

	r = NewRequest()
	r.Env["REMOTE_USER"] = "-admin"
	assert.False(t, hasAdminPriviledge(r))

	r = NewRequest()
	r.Env["REMOTE_USER"] = "testevent-admin"
	assert.False(t, hasAdminPriviledge(r))
}

func TestAuthentificator(t *testing.T) {

	// Create a test event
	repositoryService := NewRepositoryService("security.db")
	repositoryService.CreateCollectionIfNotExists(EventsBucketName)
	defer repositoryService.ShutDown()
	defer os.Remove("security.db")

	eventService := &EventService{
		RepositoryService: repositoryService,
		Secret:            "secret",
	}

	event, err := entities.NewPendingConfirmationEvent("testevent", "ip", "test@test.com")
	assert.NoError(t, err)

	eventService.ConfirmAndSaveEvent(event)
	eventAdminPass := event.AdminPassword
	assert.Len(t, eventAdminPass, 8)

	// Wrong login or pass
	assert.False(t, Authenticate(eventService, "superpass", "a login", "a pass"))
	assert.False(t, Authenticate(eventService, "superpass", "", "a pass"))
	assert.False(t, Authenticate(eventService, "superpass", "a login", ""))

	// Super admin
	assert.False(t, Authenticate(eventService, "superpass", "superadmin", "a pass"))
	assert.False(t, Authenticate(eventService, "superpass", "superadmin", ""))
	assert.True(t, Authenticate(eventService, "superpass", "superadmin", "superpass"))

	// Event admin
	assert.False(t, Authenticate(eventService, "superpass", event.Code+"-admin", "a pass"))
	assert.False(t, Authenticate(eventService, "superpass", event.Code+"-admin", ""))
	assert.True(t, Authenticate(eventService, "superpass", event.Code+"-admin", eventAdminPass))
}

func TestNewPassword(t *testing.T) {
	assert.Len(t, NewPassword(4), 4)
	assert.Len(t, NewPassword(10), 10)
}
