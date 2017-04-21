package services

import (
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
)

func TestProperties(t *testing.T) {
	repositoryService := NewRepositoryService("properties.db")
	defer repositoryService.ShutDown()
	defer os.Remove("properties.db")

	assert.NotNil(t, repositoryService)
	repositoryService.CreateCollectionIfNotExists(PropertiesBucketName)

	assert.Equal(t, "a", GetProperty(repositoryService, "code", "a"))
	assert.Equal(t, "a", GetProperty(repositoryService, "code", "b"))
}
