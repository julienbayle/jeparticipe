package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	StateOpen   = "open"
	StateClosed = "close"
)

type Activity struct {
	Code         string
	State        string
	Participants []*Participant
}

type Participant struct {
	Code        string    `json:"code"`
	PublicText  string    `json:"text"`
	PrivateText string    `json:"admintext"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
	DeletedAt   time.Time `json:"deletedAt"`
}

// Creates a new activity
func NewActivity(code string) *Activity {
	return &Activity{
		Code:         code,
		State:        StateOpen,
		Participants: make([]*Participant, 0),
	}
}

// Adds a participant to an activity
// A new participant will be auto-deleted one year later
func (activity *Activity) AddParticipant(publicText string, privateText string, ip string) *Participant {
	p := &Participant{
		PublicText:  publicText,
		PrivateText: privateText,
		CreatedAt:   time.Now(),
		CreatedBy:   ip,
		DeletedAt:   time.Now().AddDate(1, 0, 0),
	}
	p.Code = generateParticipantCode(p)
	activity.Participants = append(activity.Participants, p)

	return p
}

// Returns a participant from an activity
func (activity *Activity) GetParticipant(code string) *Participant {
	for k, v := range activity.Participants {
		if v.Code == code {
			return activity.Participants[k]
		}
	}
	return nil
}

// Flags a participant as deleted
func (activity *Activity) RemoveParticipant(code string) *Participant {
	p := activity.GetParticipant(code)
	if p != nil {
		p.DeletedAt = time.Now()
	}
	return p
}

// Removes activity data that should not been seen by non-admin users or users with another IP
func (activity *Activity) RemovePrivateData(ip string) {
	filteredParticipants := make([]*Participant, 0)
	for _, participant := range activity.Participants {
		if participant.DeletedAt.After(time.Now()) {
			if ip != participant.CreatedBy {
				participant.CreatedBy = ""
				participant.PrivateText = ""
			}
			filteredParticipants = append(filteredParticipants, participant)
		}
	}
	activity.Participants = filteredParticipants
}

// Returns if the state field has a valid value
func (activity *Activity) IsStateValid() bool {
	s := activity.State
	return s == StateOpen || s == StateClosed
}

// Returns true if state equals to "open"
func (activity *Activity) IsOpen() bool {
	return activity.State == StateOpen
}

// Computes a participant code using a hash
func generateParticipantCode(p *Participant) string {
	h := sha256.New()
	h.Write([]byte(p.PublicText + p.PrivateText + p.CreatedBy + p.CreatedAt.Format(time.RFC3339Nano)))
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes[:])
}
