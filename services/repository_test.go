package services_test

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/julienbayle/jeparticipe/app/test"
	"github.com/julienbayle/jeparticipe/services"
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
)

type testData struct {
	Field1 string
	field2 string
}

func TestBoltRepository(t *testing.T) {
	repositoryService := services.NewRepositoryService("repo.db")
	defer repositoryService.ShutDown()
	defer os.Remove("repo.db")

	assert.NotNil(t, repositoryService)

	repositoryService.CreateCollectionIfNotExists("testcollection")

	data := &testData{}
	assert.Nil(t, repositoryService.GetDocument("testcollection", "testid", data))

	data.Field1 = "Field1"
	data.field2 = "field2"

	assert.Nil(t, repositoryService.CommitDocument("testcollection", "testid", data))

	recoverData := &testData{}
	assert.Nil(t, repositoryService.GetDocument("testcollection", "testid", recoverData))
	assert.Equal(t, data.Field1, recoverData.Field1)
	assert.Equal(t, "", recoverData.field2)
}

func TestBackup(t *testing.T) {
	jeparticipe, handler, event := apptest.CreateATestApp()
	defer apptest.DeleteTestApp(jeparticipe)

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "/backup", nil))
	recorded.CodeIs(403)
	recorded.BodyIs("{\"Error\":\"Forbidden\"}")

	token := apptest.GetAdminTokenForEvent(t, &handler, event)
	rq := apptest.MakeAdminRequest("GET", "/backup", nil, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(403)
	recorded.BodyIs("{\"Error\":\"Forbidden\"}")

	token = apptest.GetSuperAdminToken(t, &handler, jeparticipe)
	rq = apptest.MakeAdminRequest("GET", "/backup", nil, token)
	recorded = test.RunRequest(t, handler, rq)
	recorded.CodeIs(200)
	recorded.HeaderIs("Content-Type", "application/octet-stream")
}
