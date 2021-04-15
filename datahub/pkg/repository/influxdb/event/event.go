package event

import (
	EntityInfluxEvent "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/event"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"time"
)

var (
	scope = log.RegisterScope("event_db_measurement", "event DB measurement", 0)
)

type EventRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewEventRepository(influxDBCfg *InternalInflux.Config) *EventRepository {
	scope.Infof("influxdb-NewEventRepository input %v", influxDBCfg)
	return &EventRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (e *EventRepository) CreateEvents(in *datahub_v1alpha1.CreateEventsRequest) error {
	scope.Infof("influxdb-CreateEvents input %v", in)
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
		pt, err := InfluxClient.NewPoint(string(Event), tags, fields, tempTime)
		if err != nil {
			scope.Error(err.Error())
		}

		points = append(points, pt)
	}

	err := e.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Event),
	})

	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (e *EventRepository) ListEvents(in *datahub_v1alpha1.ListEventsRequest) ([]*datahub_v1alpha1.Event, error) {
	scope.Infof("influxdb-ListEvents input %v", in)
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
		Measurement:    Event,
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

	results, err := e.influxDB.QueryDB(cmd, string(RepoInflux.Event))
	if err != nil {
		scope.Errorf("influxdb-ListEvents error %v", err)
		return make([]*datahub_v1alpha1.Event, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	events := e.getEventsFromInfluxRows(influxdbRows)

	scope.Infof("influxdb-ListEvents return %d %v", len(events), events)
	return events, nil
}

func (e *EventRepository) getEventsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*datahub_v1alpha1.Event {
	events := make([]*datahub_v1alpha1.Event, 0)

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

			eventType := datahub_v1alpha1.EventType_EVENT_TYPE_UNDEFINED
			if tempType, exist := data[EntityInfluxEvent.EventType]; exist {
				if value, ok := datahub_v1alpha1.EventType_value[tempType]; ok {
					eventType = datahub_v1alpha1.EventType(value)
				}
			}

			eventVersion := datahub_v1alpha1.EventVersion_EVENT_VERSION_UNDEFINED
			if tempVersion, exist := data[EntityInfluxEvent.EventVersion]; exist {
				if value, ok := datahub_v1alpha1.EventVersion_value[tempVersion]; ok {
					eventVersion = datahub_v1alpha1.EventVersion(value)
				}
			}

			eventLevel := datahub_v1alpha1.EventLevel_EVENT_LEVEL_UNDEFINED
			if tempLevel, exist := data[EntityInfluxEvent.EventLevel]; exist {
				if value, ok := datahub_v1alpha1.EventLevel_value[tempLevel]; ok {
					eventLevel = datahub_v1alpha1.EventLevel(value)
				}
			}

			event := datahub_v1alpha1.Event{
				Time:      tempTime,
				Id:        id,
				ClusterId: clusterId,
				Source: &datahub_v1alpha1.EventSource{
					Host:      sourceHost,
					Component: sourceComponent,
				},
				Type:    eventType,
				Version: eventVersion,
				Level:   eventLevel,
				Subject: &datahub_v1alpha1.K8SObjectReference{
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
