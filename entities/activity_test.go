package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensure activity constructor works
func TestActivityBasic(t *testing.T) {
	activity := NewActivity("code_test")

	assert.True(t, activity.IsOpen(), "Wrong activity initial state, should be open")
	assert.Equal(t, "code_test", activity.Code, "Wrong activity code")
	assert.Len(t, activity.Participants, 0)
}

// Ensure adding to getting participants works
func TestAddingAndGettingActivity(t *testing.T) {
	activity := NewActivity("code_test")
	p0 := activity.AddParticipant("some public text 0", "some private text 0", "IP 0")
	p1 := activity.AddParticipant("some public text 1", "some private text 1", "IP 1")
	p2 := activity.AddParticipant("some public text 2", "some private text 2", "IP")
	p3 := activity.AddParticipant("some public text 2", "some private text 2", "IP")

	assert.Len(t, activity.Participants, 4)
	assert.NotEqual(t, p0.Code, p1.Code)
	assert.NotEqual(t, p1.Code, p2.Code)
	assert.NotEqual(t, p2.Code, p3.Code)

	invalid_participant := activity.GetParticipant("invalid_code")
	assert.Nil(t, invalid_participant)

	participant := activity.GetParticipant(p2.Code)
	assert.NotNil(t, participant)
	assert.Equal(t, generateParticipantCode(participant), participant.Code)
	assert.Equal(t, "some public text 2", participant.PublicText)
	assert.Equal(t, "some private text 2", participant.PrivateText)
	assert.Equal(t, "IP", participant.CreatedBy)
	assert.True(t, participant.CreatedAt.Before(time.Now()))
	assert.True(t, participant.DeletedAt.After(time.Now()))
}

// Ensure remove participant words
func TestRemoveParticipant(t *testing.T) {
	activity := NewActivity("code_test")
	p := activity.AddParticipant("some public text", "some private text", "IP")

	assert.Nil(t, activity.RemoveParticipant("invalid_code"))
	assert.NotNil(t, activity.RemoveParticipant(p.Code))

	participant := activity.GetParticipant(p.Code)
	assert.True(t, participant.DeletedAt.Before(time.Now()))
}

// Ensure deleted participants are hidden by RemovePrivateData method
func TestRemovePrivateData_deleted(t *testing.T) {
	activity := NewActivity("code_test")
	p := activity.AddParticipant("some public text", "some private text", "IP")
	activity.RemoveParticipant(p.Code)
	activity.RemovePrivateData("IP")

	assert.Len(t, activity.Participants, 0, "Wrong participant count, only admin could see deleted participants")
}

// Ensure participant see its data (same IP)
func TestRemovePrivateData_sameIP(t *testing.T) {
	activity := NewActivity("code_test")
	p := activity.AddParticipant("some public text", "some private text", "IP")
	activity.RemovePrivateData("IP")

	assert.Len(t, activity.Participants, 1)

	participant := activity.GetParticipant(p.Code)
	assert.Equal(t, "some public text", participant.PublicText)
	assert.Equal(t, "some private text", participant.PrivateText)
	assert.Equal(t, "IP", participant.CreatedBy)
}

// Ensure a participant with a different IP can't see other private data
func TestRemovePrivateData_differentIP(t *testing.T) {
	activity := NewActivity("code_test")
	p := activity.AddParticipant("some public text", "some private text", "IP")
	activity.RemovePrivateData("other IP")

	assert.Len(t, activity.Participants, 1)

	participant := activity.GetParticipant(p.Code)
	assert.Equal(t, "some public text", participant.PublicText)
	assert.Equal(t, "", participant.PrivateText)
	assert.Equal(t, "", participant.CreatedBy)
}

// Ensure state validation works
func TestStateValidation(t *testing.T) {
	activity := NewActivity("code_test")
	activity.State = "other"
	assert.False(t, activity.IsStateValid())
}
