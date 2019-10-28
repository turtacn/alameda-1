package events

type eventTag = string
type eventField = string

const (
	EventTime              eventTag = "time"
	EventClusterId         eventTag = "cluster_id"
	EventSourceComponent   eventTag = "source_component"
	EventSourceHost        eventTag = "source_host"
	EventType              eventTag = "type"
	EventVersion           eventTag = "version"
	EventLevel             eventTag = "level"
	EventSubjectKind       eventTag = "subject_kind"
	EventSubjectNamespace  eventTag = "subject_namespace"
	EventSubjectName       eventTag = "subject_name"
	EventSubjectApiVersion eventTag = "subject_api_version"

	EventId      eventField = "id"
	EventMessage eventField = "message"
	EventData    eventField = "data"
)

var (
	// ControllerTags is list of tags of alameda_controller_recommendation measurement
	EventTags = []eventTag{
		EventTime,
		EventClusterId,
		EventSourceComponent,
		EventSourceHost,
		EventType,
		EventVersion,
		EventLevel,
		EventSubjectKind,
		EventSubjectNamespace,
		EventSubjectName,
		EventSubjectApiVersion,
	}
	// ControllerFields is list of fields of alameda_controller_recommendation measurement
	EventFields = []eventField{
		EventId,
		EventMessage,
		EventData,
	}
)
