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

// Returns an activity by its code or inits a new activity without saving it to the database
func (as *ActivityService) GetActivity(w rest.ResponseWriter, r *rest.Request) {
	activity := as.getOrCreateActivityFromRequest(r)
	returnActivityAsJson(activity, w, r)
}

// Adds a participant to an activity
func (as *ActivityService) AddAParticipantToAnActivity(w rest.ResponseWriter, r *rest.Request) {
	activity := as.getOrCreateActivityFromRequest(r)

	if !activity.IsOpen() && !hasAdminPriviledge(r) {
		rest.Error(w, "Access forbidden", http.StatusForbidden)
		return
	}

	if r.ContentLength > 512 {
		rest.Error(w, "Participant data is limited to 512 characters.", http.StatusNotAcceptable)
		return
	}

	if len(activity.Participants) > 100 {
		rest.Error(w, "Number of participants has reach the limit", http.StatusUnauthorized)
		return
	}

	participant := &entities.Participant{}
	err := r.DecodeJsonPayload(&participant)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	if participant.PublicText == "" {
		rest.Error(w, "Some public text required", http.StatusNotAcceptable)
		return
	}

	activity.AddParticipant(participant.PublicText, participant.PrivateText, getIp(r))

	as.SaveActivity(activity, getEventCodeFromRequest(r))
	returnActivityAsJson(activity, w, r)
}

// Removes a participant from an activity
func (as *ActivityService) RemoveAParticipantFromAnActivity(w rest.ResponseWriter, r *rest.Request) {
	activity := as.getOrCreateActivityFromRequest(r)
	participant := activity.GetParticipant(getParticipantCodeFromRequest(r))

	if participant == nil {
		rest.NotFound(w, r)
		return
	}

	if (getIp(r) != participant.CreatedBy || !activity.IsOpen()) && !hasAdminPriviledge(r) {
		rest.Error(w, "Access forbidden", http.StatusForbidden)
		return
	}

	activity.RemoveParticipant(participant.Code)
	as.SaveActivity(activity, getEventCodeFromRequest(r))
	returnActivityAsJson(activity, w, r)
}

// Updates activity state
func (as *ActivityService) UpdateActivityState(w rest.ResponseWriter, r *rest.Request) {
	activity := as.getOrCreateActivityFromRequest(r)
	activity.State = r.PathParam("state")

	if !activity.IsStateValid() {
		rest.Error(w, "Invalid state", http.StatusNotAcceptable)
		return
	}

	if !hasAdminPriviledge(r) {
		rest.Error(w, "Access forbidden", http.StatusForbidden)
		return
	}

	as.SaveActivity(activity, getEventCodeFromRequest(r))
	returnActivityAsJson(activity, w, r)
}

// Gets an activity from bolt database or creates a new one (without saving it to the database)
func (as *ActivityService) GetOrCreateActivity(activityCode string, eventCode string) *entities.Activity {
	activity := entities.NewActivity(activityCode)
	as.RepositoryService.GetDocument(GetActivityBucketName(eventCode), activity.Code, activity)
	return activity
}

// Saves an activity to the bolt database
func (as *ActivityService) SaveActivity(activity *entities.Activity, eventCode string) error {
	if !activity.IsStateValid() {
		return errors.New("Activity can't be saved, invalid state")
	}
	return as.RepositoryService.CommitDocument(GetActivityBucketName(eventCode), activity.Code, activity)
}

// Convenient method to get current activity using request parameters as criteria
func (as *ActivityService) getOrCreateActivityFromRequest(r *rest.Request) *entities.Activity {
	activityCode := getActivityCodeFromRequest(r)
	eventCode := getEventCodeFromRequest(r)
	return as.GetOrCreateActivity(activityCode, eventCode)
}

// Convenient method to generate bucket name for one specific event
func GetActivityBucketName(eventCode string) string {
	return "activities" + "-" + eventCode
}

// Convenient method to not forget to remove private data if needed when sendint back activiy
func returnActivityAsJson(activity *entities.Activity, w rest.ResponseWriter, r *rest.Request) {

	// Hides private information if user does not have admin priviledges
	if !hasAdminPriviledge(r) {
		activity.RemovePrivateData(getIp(r))
	}
	w.WriteJson(activity)
}

// Convenient method to get activity code from request
func getActivityCodeFromRequest(r *rest.Request) string {
	extractor, _ := regexp.Compile("[-A-Za-z0-9]{2,50}")
	return extractor.FindString(r.PathParam("acode"))
}

// Convenient method to get participant code from request
func getParticipantCodeFromRequest(r *rest.Request) string {
	extractor, _ := regexp.Compile("[a-f0-9]{64}")
	return extractor.FindString(r.PathParam("pcode"))
}
