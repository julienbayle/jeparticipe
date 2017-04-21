package app_test

import (
	"github.com/julienbayle/jeparticipe/app/test"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestApp(t *testing.T) {
	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	token := apptest.GetAdminTokenForEvent(t, &handler, event)
	assert.NotEmpty(t, token)
	recorder := apptest.MakeAdminRequest(t, &handler, "PUT", "/event/"+event.Code+"/activity/test/state/close", nil, token)
	recorder.CodeIs(200)
}
