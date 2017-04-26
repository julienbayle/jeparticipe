package services_test

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/julienbayle/jeparticipe/app/test"
	"github.com/julienbayle/jeparticipe/entities"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestGetAndAddParticipantActivityService(t *testing.T) {

	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Get a new activity (Event does not exist)
	// ------------------------------------

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/donotexistsevent/activity/testbucket", nil))
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Invalid event code\"}")

	// ------------------------------------
	// Get a new activity (Event not confirmed)
	// ------------------------------------

	eventNotConfirmed, _ := entities.NewPendingConfirmationEvent("notconfirmed", "ip", "test@test.com")
	jeparticipe.EventService.SaveEvent(eventNotConfirmed)
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/notconfirmed/activity/testbucket", nil))
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Event not confirmed yet\"}")

	// ------------------------------------
	// Get a new activity
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/testevent/activity/testbucket", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity := &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.True(t, activity.IsOpen(), "Wrong activity initial state, should be open")
	assert.Equal(t, "testbucket", activity.Code, "Wrong activity code")
	assert.Len(t, activity.Participants, 0)

	// ------------------------------------
	// Add a participant
	// ------------------------------------

	data := &map[string]string{"text": "public", "admintext": "private"}
	rq := test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/participant", data)
	rq.Header.Set("X-Real-IP", "12.12.12.12")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)
	participant := activity.Participants[0]

	// ------------------------------------
	// Get the activity back (same user)
	// ------------------------------------

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testbucket", nil)
	rq.Header.Set("X-Real-IP", "12.12.12.12")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)
	assert.Equal(t, "testbucket", activity.Code)
	assert.Equal(t, "private", activity.GetParticipant(participant.Code).PrivateText)
	assert.Equal(t, "12.12.12.12", activity.GetParticipant(participant.Code).CreatedBy)

	// ------------------------------------
	// Get the activity back (other user)
	// ------------------------------------

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testbucket", nil)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)
	assert.Equal(t, "testbucket", activity.Code)
	assert.Empty(t, activity.GetParticipant(participant.Code).PrivateText)
	assert.Empty(t, activity.GetParticipant(participant.Code).CreatedBy)

	// ------------------------------------
	// Get the activity back (admin)
	// ------------------------------------

	token := apptest.GetAdminTokenForEvent(t, &handler, event)
	rq = apptest.MakeAdminRequest("GET", "/event/testevent/activity/testbucket", nil, token)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)
	assert.Equal(t, "testbucket", activity.Code)
	assert.Equal(t, "private", activity.GetParticipant(participant.Code).PrivateText)
	assert.Equal(t, "12.12.12.12", activity.GetParticipant(participant.Code).CreatedBy)

	// ------------------------------------
	// Add many participants (with same data)
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/participant", data))
	recorded.CodeIs(200)
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/participant", data))
	recorded.CodeIs(200)
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/participant", data))
	recorded.CodeIs(200)

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 4)

	// ------------------------------------
	// Add participant with too much data
	// ------------------------------------

	longtext := "a"
	for i := 0; i < 500; i++ {
		longtext = longtext + "b"
	}
	data = &map[string]string{"text": longtext, "admintext": "private"}
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/participant", data))
	recorded.CodeIs(406)

	// ------------------------------------
	// Add too much participants
	// ------------------------------------

	data = &map[string]string{"text": "public", "admintext": "private"}
	for i := 0; i <= 100; i++ {
		recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testlimit/participant", data))
		recorded.CodeIs(200)
	}
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testlimit/participant", data))
	recorded.CodeIs(401)

	// ------------------------------------
	// Add participant with no public data
	// ------------------------------------

	data = &map[string]string{"text": "", "admintext": "private"}
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/participant", data))
	recorded.CodeIs(406)

	// ------------------------------------
	// Add participant to a closed activity as user
	// ------------------------------------

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testclose", event.Code)
	activity.State = "close"
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	data = &map[string]string{"text": "public", "admintext": "private"}
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testclose/participant", data))
	recorded.CodeIs(403)

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testclose", event.Code)
	assert.Len(t, activity.Participants, 0)

	// ------------------------------------
	// Add participant to a closed activity as admin
	// ------------------------------------

	rq = apptest.MakeAdminRequest("PUT", "/event/testevent/activity/testclose/participant", data, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testclose", event.Code)
	assert.Len(t, activity.Participants, 1)
}

func TestRemoveParticipantActivityService(t *testing.T) {

	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Remove a participant that does not exist
	// ------------------------------------

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/testevent/activity/testbucket/participant/testcode/delete", nil))
	recorded.CodeIs(404)

	// ------------------------------------
	// Remove a participant from a non existing event
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/event/donotexistsevent/activity/testbucket/participant/testcode/delete", nil))
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Invalid event code\"}")

	// ------------------------------------
	// Remove a participant with success (same IP / open activity)
	// ------------------------------------

	activity := jeparticipe.ActivityService.GetOrCreateActivity("testremove", event.Code)
	participant := activity.AddParticipant("public", "private", "111.111.111.111")
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	rq := test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove/participant/"+participant.Code+"/delete", nil)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove", nil)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 0)

	// ------------------------------------
	// Remove the participant (other IP / open activity)
	// ------------------------------------

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testremove2", event.Code)
	participant = activity.AddParticipant("public", "private", "111.111.111.111")
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove2/participant/"+participant.Code+"/delete", nil)
	rq.Header.Set("X-Real-IP", "22.22.22.22")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(403)

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove2", nil)
	rq.Header.Set("X-Real-IP", "22.22.22.22")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)

	// ------------------------------------
	// Remove the participant (admin / open activity)
	// ------------------------------------

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testremove3", event.Code)
	participant = activity.AddParticipant("public", "private", "111.111.111.111")
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	token := apptest.GetAdminTokenForEvent(t, &handler, event)
	rq = apptest.MakeAdminRequest("GET", "/event/testevent/activity/testremove3/participant/"+participant.Code+"/delete", nil, token)
	rq.Header.Set("X-Real-IP", "33.33.33.33")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove3", nil)
	rq.Header.Set("X-Real-IP", "22.22.22.22")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 0)

	// ------------------------------------
	// Remove a participant with success (same IP / closed activity)
	// ------------------------------------

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testremove4", event.Code)
	participant = activity.AddParticipant("public", "private", "111.111.111.111")
	activity.State = "close"
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove4/participant/"+participant.Code+"/delete", nil)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(403)

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove4", nil)
	rq.Header.Set("X-Real-IP", "111.111.111.111")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)

	// ------------------------------------
	// Remove the participant (other IP / closed activity)
	// ------------------------------------

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testremove5", event.Code)
	participant = activity.AddParticipant("public", "private", "111.111.111.111")
	activity.State = "close"
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove5/participant/"+participant.Code+"/delete", nil)
	rq.Header.Set("X-Real-IP", "22.22.22.22")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(403)

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove5", nil)
	rq.Header.Set("X-Real-IP", "22.22.22.22")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 1)

	// ------------------------------------
	// Remove the participant (admin / open activity)
	// ------------------------------------

	activity = jeparticipe.ActivityService.GetOrCreateActivity("testremove6", event.Code)
	participant = activity.AddParticipant("public", "private", "111.111.111.111")
	activity.State = "close"
	assert.NoError(t, jeparticipe.ActivityService.SaveActivity(activity, event.Code))

	rq = apptest.MakeAdminRequest("GET", "/event/testevent/activity/testremove6/participant/"+participant.Code+"/delete", nil, token)
	rq.Header.Set("X-Real-IP", "33.33.33.33")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)

	rq = test.MakeSimpleRequest("GET", "/event/testevent/activity/testremove6", nil)
	rq.Header.Set("X-Real-IP", "22.22.22.22")
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	activity = &entities.Activity{}
	assert.NoError(t, recorded.DecodeJsonPayload(&activity))
	assert.Len(t, activity.Participants, 0)
}

func TestChangeStateActivityService(t *testing.T) {

	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	// ------------------------------------
	// Change bucket state (Event does not exist)
	// ------------------------------------

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/donotexists/activity/testbucket/state/test", nil))
	recorded.CodeIs(404)
	recorded.BodyIs("{\"Error\":\"Invalid event code\"}")

	// ------------------------------------
	// Change bucket state (invalid state)
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/state/test", nil))
	recorded.CodeIs(406)

	// ------------------------------------
	// Change bucket state without permission
	// ------------------------------------

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "/event/testevent/activity/testbucket/state/close", nil))
	recorded.CodeIs(403)

	// ------------------------------------
	// Log on as event admin and change bucket state
	// ------------------------------------

	token := apptest.GetAdminTokenForEvent(t, &handler, event)
	rq := apptest.MakeAdminRequest("PUT", "/event/testevent/activity/testbucket/state/close", nil, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)

	activity := jeparticipe.ActivityService.GetOrCreateActivity("testbucket", "testevent")
	assert.Equal(t, entities.StateClosed, activity.State)
}
