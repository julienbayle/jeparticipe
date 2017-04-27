package services_test

import (
	"github.com/ant0ine/go-json-rest/rest"
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

type ComplexJson struct {
	Text  string
	Child *ComplexJson
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

func TestGetEventConfig(t *testing.T) {
	jeparticipe, handler, _ := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Event does not exist
	// ------------------------------------

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/invalidcode/config", nil))
	recorded.CodeIs(404)

	// ------------------------------------
	// Event exists but is not confirmed yet
	// ------------------------------------

	event, _ := entities.NewPendingConfirmationEvent("notconfirmed", "ip", "test@test.com")
	jeparticipe.EventService.SaveEvent(event)

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/notconfirmed/config", nil))
	recorded.CodeIs(400)

	// ------------------------------------
	// Event exists and is confirmed but config is not initialized yet
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/testevent/config", nil))
	recorded.CodeIs(200)
	recorded.BodyIs("{}")

	// ------------------------------------
	// Event exists and is confirmed, config is an invalid JSON (quote missing)
	// ------------------------------------

	assert.Panics(t, func() {
		rq := test.MakeSimpleRequest("GET", "/event/badjson/config", nil)
		jeparticipe.EventService.GetEventConfig(nil, &rest.Request{Request: rq, PathParams: map[string]string{"event": "badjson"}, Env: nil})
	})

	// ------------------------------------
	// Event exists and is confirmed, config is a valid JSON
	// ------------------------------------

	eventGoodJson, _ := entities.NewPendingConfirmationEvent("goodjson", "ip", "test@test.com")
	eventGoodJson.Config = []byte(`{"test":"test2", "test3":{"test4":5, "test6":5.0, "test7": [5, 2, 1], "test8": [], "test9":{}}}`)
	jeparticipe.EventService.ConfirmAndSaveEvent(eventGoodJson)

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/goodjson/config", nil))
	recorded.CodeIs(200)
	recorded.BodyIs(`{"test":"test2","test3":{"test4":5,"test6":5,"test7":[5,2,1],"test8":[],"test9":{}}}`)
}

func TestSetEventConfig(t *testing.T) {
	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Event does not exists
	// ------------------------------------

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/invalidcode/config", nil))
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Invalid code\"}")

	// ------------------------------------
	// Event exists but is not confirmed yet
	// ------------------------------------

	eventNotConfirmed, _ := entities.NewPendingConfirmationEvent("notconfirmed", "ip", "test@test.com")
	jeparticipe.EventService.SaveEvent(eventNotConfirmed)

	token := apptest.GetAdminTokenForEvent(t, &handler, eventNotConfirmed)
	rq := apptest.MakeAdminRequest("PUT", "/event/notconfirmed/config", nil, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(400)
	recorded.BodyIs("{\"Error\":\"Event not confirmed yet\"}")

	// ------------------------------------
	// Event exists and is confirmed / Send a valid JSON as config as an anonymous user
	// ------------------------------------

	data := &ComplexJson{
		Text: "toto",
		Child: &ComplexJson{
			Text: "toto2",
		},
	}
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/config", data))
	recorded.CodeIs(403)
	recorded.BodyIs(`{"Error":"Access forbidden"}`)

	// ------------------------------------
	// Event exists and is confirmed / Send a valid JSON as config as admin
	// ------------------------------------

	token = apptest.GetAdminTokenForEvent(t, &handler, event)
	rq = apptest.MakeAdminRequest("PUT", "/event/testevent/config", data, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.BodyIs("")

	eventModified := jeparticipe.EventService.GetEvent(event.Code)
	assert.Equal(t, `{"Text":"toto","Child":{"Text":"toto2","Child":null}}`, string(eventModified.Config))

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/testevent/config", nil))
	recorded.CodeIs(200)
	recorded.BodyIs(`{"Child":{"Child":null,"Text":"toto2"},"Text":"toto"}`)

	// ------------------------------------
	// Event exists and is confirmed / Send a bad JSON as config
	// ------------------------------------

	dataBadJson := "{test:"
	rq = apptest.MakeAdminRequest("PUT", "/event/testevent/config", dataBadJson, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(400)
	recorded.BodyIs(`{"Error":"Not a valid JSON document"}`)

	// ------------------------------------
	// Event exists and is confirmed / Message is too long
	// ------------------------------------

	toolong := ""
	for i := 0; i < 50000+1; i++ {
		toolong = toolong + "m"
	}

	rq = apptest.MakeAdminRequest("PUT", "/event/testevent/config", toolong, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(400)
	recorded.BodyIs(`{"Error":"Config data size is too large (should be less than 50ko)"}`)
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

	// ------------------------------------
	// Event already exists
	// ------------------------------------

	jeparticipe.EventService.EmailRelay = &email.EmailRelay{
		Send: func(email *email.Email) error {
			t.Errorf("No email should be sent")
			return nil
		},
	}
	data = &map[string]string{"code": "myevent", "userEmail": "test@test.com"}
	rq = test.MakeSimpleRequest("POST", "/event", data)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(403)
	recorded.BodyIs("{\"Error\":\"An event with this code already exists\"}")
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
