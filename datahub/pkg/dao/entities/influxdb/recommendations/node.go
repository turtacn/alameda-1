package recommendations

type nodeTag = string
type nodeField = string

const (
	NodeTime nodeTag = "time"
	NodeName nodeTag = "name"

	NodeValue nodeField = "value"
)

var (
	NodeTags = []nodeTag{
		NodeTime,
		NodeName,
	}

	NodeFields = []nodeField{
		NodeValue,
	}
)
