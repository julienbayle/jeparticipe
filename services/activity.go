package services

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/julienbayle/jeparticipe/entities"
)

type ActivityService struct {
	RepositoryService *RepositoryService
}

// GetActivity returns an activity by its code or inits a new activity without saving it to the database
func (as *ActivityService) GetActivity(w rest.ResponseWriter, r *rest.Request) {
	activity, err := as.getOrCreateActivityFromRequest(r)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	returnActivityAsJson(activity, w, r)
}

// AddAParticipantToAnActivity adds a participant to an activity
func (as *ActivityService) AddAParticipantToAnActivity(w rest.ResponseWriter, r *rest.Request) {
	activity, err := as.getOrCreateActivityFromRequest(r)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !activity.IsOpen() && !hasAdminPriviledge(r) {
		rest.Error(w, "Access forbidden", http.StatusForbidden)
		return
	}

	if r.ContentLength > 512 {
		rest.Error(w, "Participant data is limited to 512 characters.", http.StatusBadRequest)
		return
	}

	if len(activity.Participants) > 100 {
		rest.Error(w, "Number of participants has reach the limit", http.StatusBadRequest)
		return
	}

	participant := &entities.Participant{}
	err = r.DecodeJsonPayload(&participant)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if participant.PublicText == "" {
		rest.Error(w, "Some public text required", http.StatusBadRequest)
		return
	}

	activity.AddParticipant(participant.PublicText, participant.PrivateText, getIp(r))

	err = as.SaveActivity(activity, getEventCodeFromRequest(r))
	if err != nil {
		panic(err)
	}

	returnActivityAsJson(activity, w, r)
}

// RemoveAParticipantFromAnActivity removes a participant from an activity
func (as *ActivityService) RemoveAParticipantFromAnActivity(w rest.ResponseWriter, r *rest.Request) {
	activity, err := as.getOrCreateActivityFromRequest(r)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	participant := activity.GetParticipant(getParticipantCodeFromRequest(r))

	if participant == nil {
		rest.NotFound(w, r)
		return
	}

	if (getIp(r) != participant.CreatedBy || !activity.IsOpen()) && !hasAdminPriviledge(r) {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	activity.RemoveParticipant(participant.Code)
	err = as.SaveActivity(activity, getEventCodeFromRequest(r))
	if err != nil {
		panic(err)
	}

	returnActivityAsJson(activity, w, r)
}

// UpdateActivityState updates activity state
func (as *ActivityService) UpdateActivityState(w rest.ResponseWriter, r *rest.Request) {
	activity, err := as.getOrCreateActivityFromRequest(r)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	activity.State = r.PathParam("state")

	if !activity.IsStateValid() {
		rest.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	if !hasAdminPriviledge(r) {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = as.SaveActivity(activity, getEventCodeFromRequest(r))
	if err != nil {
		panic(err)
	}

	returnActivityAsJson(activity, w, r)
}

// GetOrCreateActivity gets an activity from bolt database or creates a new one (without saving it to the database)
func (as *ActivityService) GetOrCreateActivity(activityCode string, eventCode string) *entities.Activity {
	activity := entities.NewActivity(activityCode)
	as.RepositoryService.GetDocument(GetActivityBucketName(eventCode), activity.Code, activity)
	return activity
}

// SaveActivity saves an activity to the bolt database
func (as *ActivityService) SaveActivity(activity *entities.Activity, eventCode string) error {
	if !activity.IsStateValid() {
		return errors.New("Activity can't be saved, invalid state")
	}
	return as.RepositoryService.CommitDocument(GetActivityBucketName(eventCode), activity.Code, activity)
}

// getOrCreateActivityFromRequest is a convenient method to get current activity using request parameters as criteria
func (as *ActivityService) getOrCreateActivityFromRequest(r *rest.Request) (*entities.Activity, error) {
	activityCode := getActivityCodeFromRequest(r)
	eventCode := getEventCodeFromRequest(r)

	event := &entities.Event{}
	as.RepositoryService.GetDocument(EventsBucketName, eventCode, event)

	if event.Code == "" {
		return nil, errors.New("Invalid event code")
	}

	if !event.EmailConfirmed {
		return nil, errors.New("Event not confirmed yet")
	}

	return as.GetOrCreateActivity(activityCode, eventCode), nil
}

// GetActivityBucketName is a convenient method to generate the collection name where to save activities for a specific event
func GetActivityBucketName(eventCode string) string {
	return "activities" + "-" + eventCode
}

// returnActivityAsJson is a convenient method to not forget to remove private data if needed when sendint back activiy
func returnActivityAsJson(activity *entities.Activity, w rest.ResponseWriter, r *rest.Request) {

	// Hides private information if user does not have admin priviledges
	if !hasAdminPriviledge(r) {
		activity.RemovePrivateData(getIp(r))
	}
	w.WriteJson(activity)
}

// getActivityCodeFromRequest is a convenient method to get activity code from request
func getActivityCodeFromRequest(r *rest.Request) string {
	extractor, _ := regexp.Compile("[-A-Za-z0-9]{2,50}")
	return extractor.FindString(r.PathParam("acode"))
}

// getParticipantCodeFromRequest is a convenient method to get participant code from request
func getParticipantCodeFromRequest(r *rest.Request) string {
	extractor, _ := regexp.Compile("[a-f0-9]{64}")
	return extractor.FindString(r.PathParam("pcode"))
}
