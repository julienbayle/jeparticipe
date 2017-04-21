package app_test

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/julienbayle/jeparticipe/app/test"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestApp(t *testing.T) {
	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	token := apptest.GetAdminTokenForEvent(t, &handler, event)
	assert.NotEmpty(t, token)
	rq := apptest.MakeAdminRequest("PUT", "/event/"+event.Code+"/activity/test/state/close", nil, token)
	recorder := test.RunRequest(t, handler, rq)
	recorder.CodeIs(200)
}
