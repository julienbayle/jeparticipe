package services

import (
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
)

type testData struct {
	Field1 string
	field2 string
}

func TestBoltRepository(t *testing.T) {
	repositoryService := NewRepositoryService("repo.db")
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
