package eventmgt

import (
	"encoding/json"
	EntityInflux      "github.com/containers-ai/alameda/internal/pkg/database/entity/influxdb"
	EntityInfluxEvent "github.com/containers-ai/alameda/internal/pkg/database/entity/influxdb/event"
	DBCommon          "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux    "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalRabbitmq  "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	DatahubLog        "github.com/containers-ai/alameda/pkg/utils/log"
	ApiEvents         "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	"github.com/golang/protobuf/ptypes"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"time"
)

var (
	scope = DatahubLog.RegisterScope("event_db_measurement", "event DB measurement", 0)
)

func (e *EventMgt) PostEvents(in *ApiEvents.CreateEventsRequest) error {
	points := make([]*InfluxClient.Point, 0)

	for _, event := range in.GetEvents() {
		tags := map[string]string{
			EntityInfluxEvent.EventClusterId:         event.GetClusterId(),
			EntityInfluxEvent.EventSourceHost:        event.GetSource().GetHost(),
			EntityInfluxEvent.EventSourceComponent:   event.GetSource().GetComponent(),
			EntityInfluxEvent.EventType:              event.GetType().String(),
			EntityInfluxEvent.EventVersion:           event.GetVersion().String(),
			EntityInfluxEvent.EventLevel:             event.GetLevel().String(),
			EntityInfluxEvent.EventSubjectKind:       event.GetSubject().GetKind(),
			EntityInfluxEvent.EventSubjectNamespace:  event.GetSubject().GetNamespace(),
			EntityInfluxEvent.EventSubjectName:       event.GetSubject().GetName(),
			EntityInfluxEvent.EventSubjectApiVersion: event.GetSubject().GetApiVersion(),
		}

		fields := map[string]interface{}{
			EntityInfluxEvent.EventId:      event.GetId(),
			EntityInfluxEvent.EventMessage: event.GetMessage(),
			EntityInfluxEvent.EventData:    event.GetData(),
		}

		tempTime, _ := ptypes.Timestamp(event.GetTime())
		pt, err := InfluxClient.NewPoint(string(EntityInfluxEvent.EventMeasurement), tags, fields, tempTime)
		if err != nil {
			scope.Error(err.Error())
		}

		points = append(points, pt)
	}

	err := e.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(EntityInflux.Event),
	})

	if err != nil {
		scope.Error(err.Error())
		return err
	}

	//send to rabbitmq
	err = e.sendEventsToMsgQueue(in)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (e *EventMgt) ListEvents(in *ApiEvents.ListEventsRequest) ([]*ApiEvents.Event, error) {
	idList := in.GetId()
	clusterIdList := in.GetClusterId()

	eventTypeList := make([]string, 0)
	for _, eventType := range in.GetType() {
		eventTypeList = append(eventTypeList, eventType.String())
	}

	eventVersionList := make([]string, 0)
	for _, eventVersion := range in.GetVersion() {
		eventVersionList = append(eventVersionList, eventVersion.String())
	}

	eventLevelList := make([]string, 0)
	for _, eventLevel := range in.GetLevel() {
		eventLevelList = append(eventLevelList, eventLevel.String())
	}

	influxdbStatement := InternalInflux.Statement{
		Measurement:    EntityInfluxEvent.EventMeasurement,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	influxdbStatement.AppendWhereClauseByList(EntityInfluxEvent.EventId, "=", "OR", idList)
	influxdbStatement.AppendWhereClauseByList(EntityInfluxEvent.EventClusterId, "=", "OR", clusterIdList)
	influxdbStatement.AppendWhereClauseByList(EntityInfluxEvent.EventType, "=", "OR", eventTypeList)
	influxdbStatement.AppendWhereClauseByList(EntityInfluxEvent.EventVersion, "=", "OR", eventVersionList)
	influxdbStatement.AppendWhereClauseByList(EntityInfluxEvent.EventLevel, "=", "OR", eventLevelList)

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := e.influxDB.QueryDB(cmd, string(EntityInflux.Event))
	if err != nil {
		return make([]*ApiEvents.Event, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	events := e.getEventsFromInfluxRows(influxdbRows)

	return events, nil
}

func (e *EventMgt) getEventsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiEvents.Event {
	events := make([]*ApiEvents.Event, 0)

	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			t, _ := time.Parse(time.RFC3339Nano, data[EntityInfluxEvent.EventTime])
			tempTime, _ := ptypes.TimestampProto(t)

			clusterId := data[EntityInfluxEvent.EventClusterId]
			sourceHost := data[EntityInfluxEvent.EventSourceHost]
			sourceComponent := data[EntityInfluxEvent.EventSourceComponent]
			subjectKind := data[EntityInfluxEvent.EventSubjectKind]
			subjectNamespace := data[EntityInfluxEvent.EventSubjectNamespace]
			subjectName := data[EntityInfluxEvent.EventSubjectName]
			subjectApiVersion := data[EntityInfluxEvent.EventSubjectApiVersion]

			id := data[EntityInfluxEvent.EventId]
			message := data[EntityInfluxEvent.EventMessage]
			eventData := data[EntityInfluxEvent.EventData]

			eventType := ApiEvents.EventType_EVENT_TYPE_UNDEFINED
			if tempType, exist := data[EntityInfluxEvent.EventType]; exist {
				if value, ok := ApiEvents.EventType_value[tempType]; ok {
					eventType = ApiEvents.EventType(value)
				}
			}

			eventVersion := ApiEvents.EventVersion_EVENT_VERSION_UNDEFINED
			if tempVersion, exist := data[EntityInfluxEvent.EventVersion]; exist {
				if value, ok := ApiEvents.EventVersion_value[tempVersion]; ok {
					eventVersion = ApiEvents.EventVersion(value)
				}
			}

			eventLevel := ApiEvents.EventLevel_EVENT_LEVEL_UNDEFINED
			if tempLevel, exist := data[EntityInfluxEvent.EventLevel]; exist {
				if value, ok := ApiEvents.EventLevel_value[tempLevel]; ok {
					eventLevel = ApiEvents.EventLevel(value)
				}
			}

			event := ApiEvents.Event{
				Time:      tempTime,
				Id:        id,
				ClusterId: clusterId,
				Source: &ApiEvents.EventSource{
					Host:      sourceHost,
					Component: sourceComponent,
				},
				Type:    eventType,
				Version: eventVersion,
				Level:   eventLevel,
				Subject: &ApiEvents.K8SObjectReference{
					Kind:       subjectKind,
					Namespace:  subjectNamespace,
					Name:       subjectName,
					ApiVersion: subjectApiVersion,
				},
				Message: message,
				Data:    eventData,
			}

			events = append(events, &event)
		}
	}

	return events
}

func (e *EventMgt) sendEventsToMsgQueue(in *ApiEvents.CreateEventsRequest) error {
	messageQueue, err := InternalRabbitmq.NewRabbitMQSender(e.RabbitMQConfig)
	if err != nil {
		return err
	}
	defer messageQueue.Close()

	events, err := json.Marshal(in.GetEvents())
	if err != nil {
		return err
	}

	err = messageQueue.SendJsonString("event", string(events))
	return err
}
