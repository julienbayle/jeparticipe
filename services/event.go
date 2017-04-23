package services

import (
	"net/http"
	"regexp"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/julienbayle/jeparticipe/email"
	"github.com/julienbayle/jeparticipe/entities"
)

const (
	EventsBucketName = "events"
)

type EventService struct {
	RepositoryService *RepositoryService
	EmailRelay        *email.EmailRelay
	Secret            string
}

// Returns an event state (can be used to check if an event code is used or not)
func (es *EventService) GetEventStatus(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event.Code != eventCode {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	w.WriteJson(map[string]bool{"confirmed": event.EmailConfirmed})
}

// Creates a new pending confirmation event
func (es *EventService) CreatePendingEvent(w rest.ResponseWriter, r *rest.Request) {
	eventPayload := &entities.Event{}
	err := r.DecodeJsonPayload(&eventPayload)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	event, err := entities.NewPendingConfirmationEvent(eventPayload.Code, getIp(r), eventPayload.UserEmail)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	templateData := struct {
		URL string
	}{
		URL: r.BaseUrl().String() + "/" + event.Code + "/confirm/" + event.ConfirmCode(es.Secret),
	}
	email := email.NewEmail(event.UserEmail, "Circuleo - Jeparticipe ! - Confirmation de votre email", "")
	email.AddBodyUsingTemplate("../templates/confirm.html", templateData)
	es.EmailRelay.Send(email)

	es.SaveEvent(event)
}

// Validates an event (link from a event confirmation email)
func (es *EventService) ConfirmEvent(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event.Code != eventCode {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	if event.EmailConfirmed {
		rest.Error(w, "Already confirmed", http.StatusGone)
		return
	}

	// Check validation code
	confirmCode := r.PathParam("confirm_code")
	if confirmCode != event.ConfirmCode(es.Secret) {
		rest.Error(w, "Invalid confirmation code", http.StatusForbidden)
		return
	}

	if err := es.ConfirmAndSaveEvent(event); err != nil {
		rest.Error(w, "Unable to confirm event", http.StatusInternalServerError)
		return
	}

	templateData := struct {
		URL   string
		Login string
		Pass  string
	}{
		URL:   r.BaseUrl().String() + "/" + event.Code,
		Login: GetEventAdminLogin(eventCode),
		Pass:  event.AdminPassword,
	}
	email := email.NewEmail(event.UserEmail, "Circuleo - Je participe ! - C'est parti !", "")
	email.AddBodyUsingTemplate("../templates/confirmed.html", templateData)
	es.EmailRelay.Send(email)
}

// Send an email with the event informations
func (es *EventService) SendEventInformationByMail(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event.Code != eventCode {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	if event.EmailConfirmed {
		templateData := struct {
			URL   string
			Login string
			Pass  string
			Code  string
		}{
			URL:   r.BaseUrl().String() + "/" + event.Code,
			Login: GetEventAdminLogin(eventCode),
			Pass:  event.AdminPassword,
			Code:  event.Code,
		}
		email := email.NewEmail(event.UserEmail, "Circuleo - Je participe ! - Rappel de vos informations", "")
		email.AddBodyUsingTemplate("../templates/lostaccount.html", templateData)
		es.EmailRelay.Send(email)
	} else {
		templateData := struct {
			URL string
		}{
			URL: r.BaseUrl().String() + "/" + event.Code + "/confirm/" + event.ConfirmCode(es.Secret),
		}
		email := email.NewEmail(event.UserEmail, "Circuleo - Jeparticipe ! - Confirmation de votre email", "")
		email.AddBodyUsingTemplate("../templates/confirm.html", templateData)
		es.EmailRelay.Send(email)
	}
}

// Confirms an event an init activities collection for this event
func (es *EventService) ConfirmAndSaveEvent(event *entities.Event) error {
	// Save updated event
	event.EmailConfirmed = true
	event.AdminPassword = NewPassword(8)
	es.SaveEvent(event)

	// Init activities collection for this event
	return es.RepositoryService.CreateCollectionIfNotExists(GetActivityBucketName(event.Code))
}

// Gets an event from database
func (es *EventService) GetEvent(eventCode string) *entities.Event {
	event := &entities.Event{}
	es.RepositoryService.GetDocument(EventsBucketName, eventCode, event)
	return event
}

// Saves an event to bolt database
func (es *EventService) SaveEvent(event *entities.Event) error {
	return es.RepositoryService.CommitDocument(EventsBucketName, event.Code, event)
}

// Convenient method to get event code from request
func getEventCodeFromRequest(r *rest.Request) string {
	extractor, _ := regexp.Compile("[-A-Za-z0-9]{2,50}")
	return extractor.FindString(r.PathParam("event"))
}
