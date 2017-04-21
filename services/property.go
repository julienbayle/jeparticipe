package services

const (
	PropertiesBucketName = "properties"
)

type Property struct {
	Value string
}

func GetProperty(repositoryService *RepositoryService, code string, defautValue string) string {
	prop := &Property{}
	err := repositoryService.GetDocument(PropertiesBucketName, code, prop)
	if err != nil || prop.Value == "" {
		prop.Value = defautValue
		repositoryService.CommitDocument(PropertiesBucketName, code, prop)
	}
	return prop.Value
}
