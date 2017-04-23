package services_test

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/julienbayle/jeparticipe/app/test"
	"github.com/julienbayle/jeparticipe/email"
	"github.com/julienbayle/jeparticipe/entities"
	"github.com/stretchr/testify/assert"

	"strings"
	"testing"
)

type EventState struct {
	Confirmed bool `json:"confirmed"`
}

func TestEventState(t *testing.T) {

	jeparticipe, handler, _ := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Event do not exists
	// ------------------------------------

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/invalidcode/status", nil))
	recorded.CodeIs(404)

	// ------------------------------------
	// Event exists (confirmed)
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/testevent/status", nil))
	recorded.CodeIs(200)
	eventState := &EventState{}
	assert.NoError(t, recorded.DecodeJsonPayload(&eventState))
	assert.True(t, eventState.Confirmed)

	// ------------------------------------
	// Event exists (not confirmed)
	// ------------------------------------

	event, _ := entities.NewPendingConfirmationEvent("testevent2", "ip", "test@test.com")
	jeparticipe.EventService.SaveEvent(event)

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/testevent2/status", nil))
	recorded.CodeIs(200)
	eventState = &EventState{}
	assert.NoError(t, recorded.DecodeJsonPayload(&eventState))
	assert.False(t, eventState.Confirmed)
}

func TestCreatePendingEvent(t *testing.T) {

	jeparticipe, handler, _ := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Code invalid
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			t.Errorf("No email should be sent")
			return nil
		},
	}
	data := &map[string]string{"code": "&&", "userEmail": "test@test.com"}
	rq := test.MakeSimpleRequest("POST", "/event", data)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded := test.RunRequest(t, handler, rq)
	recorded.CodeIs(406)
	recorded.BodyIs("{\"Error\":\"Invalid code\"}")

	// ------------------------------------
	// Email invalid
	// ------------------------------------

	data = &map[string]string{"code": "myevent", "userEmail": "test"}
	rq = test.MakeSimpleRequest("POST", "/event", data)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(406)
	recorded.BodyIs("{\"Error\":\"Invalid email\"}")

	// ------------------------------------
	// Valid creation
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			if !strings.Contains(email.Body, "confirm") {
				t.Errorf("Email body is suspect")
			}
			if email.To != "test@test.com" {
				t.Errorf("Bad recipient")
			}
			return nil
		},
	}
	data = &map[string]string{"code": "myevent", "userEmail": "test@test.com"}
	rq = test.MakeSimpleRequest("POST", "/event", data)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.BodyIs("")

	event := jeparticipe.EventService.GetEvent("myevent")
	assert.False(t, event.EmailConfirmed)
}

func TestConfirmEvent(t *testing.T) {

	jeparticipe, handler, _ := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Event do not exists
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			t.Errorf("No email should be sent")
			return nil
		},
	}
	rq := test.MakeSimpleRequest("GET", "/event/donotexists/confirm/badcode", nil)
	recorded := test.RunRequest(t, handler, rq)
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Invalid code\"}")

	// ------------------------------------
	// Invalid confirmation code
	// ------------------------------------

	event, _ := entities.NewPendingConfirmationEvent("myevent", "111.111.111.111", "test@test.com")
	jeparticipe.EventService.SaveEvent(event)

	rq = test.MakeSimpleRequest("GET", "/event/myevent/confirm/badcode", nil)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(403)
	recorded.BodyIs("{\"Error\":\"Invalid confirmation code\"}")

	// ------------------------------------
	// Confirmation code valid
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			updatedEvent := jeparticipe.EventService.GetEvent(event.Code)
			if !strings.Contains(email.Body, updatedEvent.AdminPassword) {
				t.Errorf("Email body is suspect : %s, password %s", email.Body, updatedEvent.AdminPassword)
			}
			if email.To != "test@test.com" {
				t.Errorf("Bad recipient")
			}
			return nil
		},
	}
	rq = test.MakeSimpleRequest("GET", "/event/myevent/confirm/"+event.ConfirmCode(jeparticipe.EventService.Secret), nil)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.BodyIs("")

	event = jeparticipe.EventService.GetEvent("myevent")
	assert.True(t, event.EmailConfirmed)

	// ------------------------------------
	// Already confirmed
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			t.Errorf("No email should be sent")
			return nil
		},
	}
	rq = test.MakeSimpleRequest("GET", "/event/myevent/confirm/"+event.ConfirmCode(jeparticipe.EventService.Secret), nil)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(410)
	recorded.BodyIs("{\"Error\":\"Already confirmed\"}")

}

func TestSendEventInformationByMail(t *testing.T) {

	jeparticipe, handler, eventConfirmed := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Event do not exists
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			t.Errorf("No email should be sent")
			return nil
		},
	}
	rq := test.MakeSimpleRequest("GET", "/event/donotexists/lostaccount", nil)
	recorded := test.RunRequest(t, handler, rq)
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Invalid code\"}")

	// ------------------------------------
	// Not confirmed
	// ------------------------------------

	eventNotConfirmed, _ := entities.NewPendingConfirmationEvent("notconfirmed", "111.111.111.111", "test@test.com")
	jeparticipe.EventService.SaveEvent(eventNotConfirmed)

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			if !strings.Contains(email.Body, eventNotConfirmed.ConfirmCode(jeparticipe.EventService.Secret)) {
				t.Errorf("Email body is suspect : %s", email.Body)
			}
			if email.To != "test@test.com" {
				t.Errorf("Bad recipient")
			}
			return nil
		},
	}
	rq = test.MakeSimpleRequest("GET", "/event/notconfirmed/lostaccount", nil)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.BodyIs("")

	// ------------------------------------
	// Confirmed
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			if !strings.Contains(email.Body, eventConfirmed.AdminPassword) {
				t.Errorf("Email body is suspect : %s", email.Body)
			}
			if email.To != "test@test.com" {
				t.Errorf("Bad recipient")
			}
			return nil
		},
	}
	rq = test.MakeSimpleRequest("GET", "/event/testevent/lostaccount", nil)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.BodyIs("")
}
