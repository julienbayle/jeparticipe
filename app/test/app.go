package apptest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/julienbayle/jeparticipe/app"
	"github.com/julienbayle/jeparticipe/entities"
	"github.com/julienbayle/jeparticipe/services"

	"net/http"
	"os"
	"testing"
)

// Creates a test application
func CreateATestApp() (*app.App, http.Handler, *entities.Event) {
	// Initialize the app
	jeparticipe := app.NewApp("test.db")

	// Initialize the API endpoint
	restapi := jeparticipe.BuildApi(app.TestMode, "")
	handler := restapi.MakeHandler()

	// Create a new event
	event, _ := entities.NewPendingConfirmationEvent("testevent", "ip", "test@test.com")
	jeparticipe.EventService.ConfirmAndSaveEvent(event)

	return jeparticipe, handler, event
}

// Closes database and remove database file
func DeleteTestApp(aApp *app.App) {
	defer aApp.ShutDown()
	defer os.Remove("test.db")
}

// Returns a login token
func GetAdminTokenForEvent(t *testing.T, handler *http.Handler, event *entities.Event) string {
	loginCreds := map[string]string{"username": event.Code + "-admin", "password": event.AdminPassword}
	rightCredReq := test.MakeSimpleRequest("POST", "/login", loginCreds)
	recorded := test.RunRequest(t, *handler, rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	nToken := services.SecurityToken{}
	test.DecodeJsonPayload(recorded.Recorder, &nToken)
	return nToken.Token
}

// Sends a request to the app api as admin
func MakeAdminRequest(method string, url string, data interface{}, token string) *http.Request {
	stateReq := test.MakeSimpleRequest(method, url, data)
	stateReq.Header.Set("Authorization", "Bearer "+token)
	return stateReq
}
