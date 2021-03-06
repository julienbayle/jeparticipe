package services

import (
	"encoding/json"
	"io/ioutil"
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

// GetEventStatus returns an event state (can be used to check if an event code is used or not)
func (es *EventService) GetEventStatus(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event == nil {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	w.WriteJson(map[string]bool{"confirmed": event.EmailConfirmed})
}

// GetEventConfig returns the config field value
func (es *EventService) GetEventConfig(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event == nil {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	if !event.EmailConfirmed {
		rest.Error(w, "Event not confirmed yet", http.StatusBadRequest)
		return
	}

	if event.Config == nil {
		w.WriteJson(map[string]string{})
		return
	}

	d := &map[string]interface{}{}
	err := json.Unmarshal(event.Config, d)
	if err != nil {
		panic("Invalid event config for " + event.Code + ", not a valid JSON file")
	}

	w.WriteJson(d)
}

// SetEventConfig updates the config field
func (es *EventService) SetEventConfig(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event == nil {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	if !hasAdminPriviledge(r) {
		rest.Error(w, "Access forbidden", http.StatusForbidden)
		return
	}

	if !event.EmailConfirmed {
		rest.Error(w, "Event not confirmed yet", http.StatusBadRequest)
		return
	}

	if r.ContentLength > 50000 {
		rest.Error(w, "Config data size is too large (should be less than 50ko)", http.StatusBadRequest)
		return
	}

	config, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event.Config = config

	d := &map[string]interface{}{}
	err = json.Unmarshal(event.Config, d)
	if err != nil {
		rest.Error(w, "Not a valid JSON document", http.StatusBadRequest)
		return
	}

	err = es.SaveEvent(event)
	if err != nil {
		panic(err)
		return
	}
}

// CreatePendingEvent creates a new pending confirmation event
func (es *EventService) CreatePendingEvent(w rest.ResponseWriter, r *rest.Request) {
	eventPayload := &entities.Event{}
	err := r.DecodeJsonPayload(&eventPayload)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	eventExist := es.GetEvent(eventPayload.Code)
	if eventExist != nil {
		rest.Error(w, "An event with this code already exists", http.StatusForbidden)
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

	err = es.SaveEvent(event)
	if err != nil {
		panic(err)
		return
	}
}

// ConfirmEvent validates an event (link from a event confirmation email)
func (es *EventService) ConfirmEvent(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event == nil {
		rest.Error(w, "Invalid code", http.StatusNotFound)
		return
	}

	if event.EmailConfirmed {
		rest.Error(w, "Already confirmed", http.StatusNotModified)
		return
	}

	// Check validation code
	confirmCode := r.PathParam("confirm_code")
	if confirmCode != event.ConfirmCode(es.Secret) {
		rest.Error(w, "Invalid confirmation code", http.StatusBadRequest)
		return
	}

	if err := es.ConfirmAndSaveEvent(event); err != nil {
		panic(err)
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

// SendEventInformationByMail sends an email to the event admin with the event informations
func (es *EventService) SendEventInformationByMail(w rest.ResponseWriter, r *rest.Request) {
	eventCode := getEventCodeFromRequest(r)
	event := es.GetEvent(eventCode)

	if event == nil {
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

// ConfirmAndSaveEvent confirms an event an init activities collection for this event
func (es *EventService) ConfirmAndSaveEvent(event *entities.Event) error {
	// Save updated event
	event.EmailConfirmed = true
	event.AdminPassword = NewPassword(8)
	es.SaveEvent(event)

	// Init activities collection for this event
	return es.RepositoryService.CreateCollectionIfNotExists(GetActivityBucketName(event.Code))
}

// GetEvent gets an event from database
func (es *EventService) GetEvent(eventCode string) *entities.Event {
	event := &entities.Event{}
	es.RepositoryService.GetDocument(EventsBucketName, eventCode, event)
	if event.Code == "" {
		return nil
	}

	return event
}

// SaveEvent saves an event to the database
func (es *EventService) SaveEvent(event *entities.Event) error {
	return es.RepositoryService.CommitDocument(EventsBucketName, event.Code, event)
}

// getEventCodeFromRequest is a convenient method to get an event code from request
func getEventCodeFromRequest(r *rest.Request) string {
	extractor, _ := regexp.Compile("[-A-Za-z0-9]{2,50}")
	return extractor.FindString(r.PathParam("event"))
}
