package prometheus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	Common "github.com/containers-ai/api/common"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func ReadRawdata(config *Config, queries []*Common.Query) ([]*Common.ReadRawdata, error) {
	rawdata := make([]*Common.ReadRawdata, 0)

	prometheusClient, err := NewClient(config)
	if err != nil {
		scope.Errorf("failed to read rawdata from Prometheus: %v", err)
		return make([]*Common.ReadRawdata, 0), errors.New("failed to instance prometheus client")
	}

	for _, query := range queries {
		response := Response{}
		err := errors.New("")

		queryExpression := ""
		queryCondition := DBCommon.BuildQueryCondition(query.GetCondition())

		options := []DBCommon.Option{
			DBCommon.StartTime(queryCondition.StartTime),
			DBCommon.EndTime(queryCondition.EndTime),
			DBCommon.Timeout(queryCondition.Timeout),
			DBCommon.StepTime(queryCondition.StepTime),
			DBCommon.AggregateOverTimeFunc(queryCondition.AggregateOverTimeFunction),
		}

		opt := DBCommon.NewDefaultOptions()
		for _, option := range options {
			option(&opt)
		}

		if query.GetCondition().GetWhereClause() != "" {
			queryExpression = fmt.Sprintf("%s{%s}", query.GetTable(), query.GetCondition().GetWhereClause())
		} else {
			queryExpression = fmt.Sprintf("%s", query.GetTable())
		}

		if query.GetCondition().GetTimeRange().GetStep() != nil {
			stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
			queryExpression, err = WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
			if err != nil {
				return make([]*Common.ReadRawdata, 0), errors.New(err.Error())
			}
		}

		switch query.GetExpression() {
		case "query":
			response, err = prometheusClient.Query(queryExpression, opt.StartTime, opt.Timeout)
		case "query_range":
			response, err = prometheusClient.QueryRange(context.TODO(), queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
		default:
			response, err = prometheusClient.QueryRange(context.TODO(), queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
		}

		if err != nil {
			return make([]*Common.ReadRawdata, 0), errors.New(err.Error())
		} else if response.Status != StatusSuccess {
			scope.Errorf("receive error response from prometheus: %s", response.Error)
			return make([]*Common.ReadRawdata, 0), errors.New(response.Error)
		} else {
			readRawdata, _ := ResponseToReadRawdata(&response, query)
			rawdata = append(rawdata, readRawdata)
		}
	}

	return rawdata, nil
}

func ResponseToReadRawdata(response *Response, query *Common.Query) (*Common.ReadRawdata, error) {
	var (
		err         error
		readRawdata = Common.ReadRawdata{Query: query}
	)

	if len(response.Data.Result) == 0 {
		return &readRawdata, nil
	}

	entities, err := response.GetEntities()
	if err != nil {
		scope.Errorf("failed to transform prometheus response to read rawdata: %s", err.Error())
		return nil, errors.New("failed to get entities from prometheus response")
	}

	// Build columns
	for key := range entities[0].Labels {
		readRawdata.Columns = append(readRawdata.Columns, key)
	}
	readRawdata.Columns = append(readRawdata.Columns, "value")

	// Build groups
	for _, entity := range entities {
		group := Common.Group{}
		for _, value := range entity.Values {
			// Build rows of group
			row := Common.Row{}
			for i := 0; i < len(readRawdata.Columns)-1; i++ {
				row.Values = append(row.Values, entity.Labels[readRawdata.Columns[i]])
			}
			row.Time = &timestamp.Timestamp{Seconds: value.UnixTime.Unix()}
			row.Values = append(row.Values, value.SampleValue)
			group.Rows = append(group.Rows, &row)
		}
		readRawdata.Groups = append(readRawdata.Groups, &group)
	}

	// Append rawdata json string
	jsonStr, err := json.Marshal(response.Data)
	if err != nil {
		scope.Errorf("failed to transform prometheus response to read rawdata: %s", err.Error())
		return nil, errors.New("failed to marshal prometheus response")
	}
	readRawdata.Rawdata = string(jsonStr)

	return &readRawdata, nil
}
