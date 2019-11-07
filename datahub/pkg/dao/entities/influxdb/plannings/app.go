package plannings

type appTag = string
type appField = string

const (
	AppTime      appTag = "time"
	AppNamespace appTag = "namespace"
	AppName      appTag = "name"

	AppValue appField = "value"
)

var (
	AppTags = []appTag{
		AppTime,
		AppNamespace,
		AppName,
	}

	AppFields = []appField{
		AppValue,
	}
)
