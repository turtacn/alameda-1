package influxdb

import (
	"fmt"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	Common "github.com/containers-ai/api/common"
	"strings"
	"time"
)

type Statement struct {
	QueryCondition *DBCommon.QueryCondition
	Database       Database
	Measurement    Measurement
	SelectedFields []string
	GroupByTags    []string
	WhereClause    string
	OrderClause    string
	LimitClause    string
}

func NewStatement(query *Common.Query) *Statement {
	if query == nil {
		return &Statement{}
	}

	queryCondition := DBCommon.BuildQueryCondition(query.GetCondition())

	statement := Statement{
		QueryCondition: &queryCondition,
		Database:       Database(query.GetDatabase()),
		Measurement:    Measurement(query.GetTable()),
		SelectedFields: query.GetCondition().GetSelects(),
		GroupByTags:    query.GetCondition().GetGroups(),
		WhereClause:    query.GetCondition().GetWhereClause(),
	}

	return &statement
}

func (s *Statement) AppendWhereClause(key string, operator string, value string) {
	if value == "" {
		return
	}

	if s.WhereClause == "" {
		s.WhereClause += fmt.Sprintf("WHERE \"%s\"%s'%s' ", key, operator, value)
	} else {
		s.WhereClause += fmt.Sprintf("AND \"%s\"%s'%s' ", key, operator, value)
	}
}

func (s *Statement) AppendWhereClauseByList(key string, operator string, listOperator string, values []string) {
	if len(values) == 0 {
		return
	}

	condition := "("
	for _, value := range values {
		condition += fmt.Sprintf("\"%s\"%s'%s' %s ", key, operator, value, listOperator)
	}
	condition = strings.TrimSuffix(condition, fmt.Sprintf("%s ", listOperator))
	condition += ")"

	if s.WhereClause == "" {
		s.WhereClause += fmt.Sprintf("WHERE %s ", condition)
	} else {
		s.WhereClause += fmt.Sprintf("AND %s ", condition)
	}
}

func (s *Statement) AppendWhereClauseDirectly(condition string) {
	if condition == "" {
		return
	}

	if s.WhereClause == "" {
		s.WhereClause += fmt.Sprintf("WHERE %s ", condition)
	} else {
		s.WhereClause += fmt.Sprintf("AND %s ", condition)
	}
}

func (s *Statement) AppendWhereClauseWithTime(operator string, value int64) {
	if value == 0 {
		return
	}

	tm := time.Unix(int64(value), 0)

	if s.WhereClause == "" {
		s.WhereClause += fmt.Sprintf("WHERE time%s'%s' ", operator, tm.UTC().Format(time.RFC3339))
	} else {
		s.WhereClause += fmt.Sprintf("AND time%s'%s' ", operator, tm.UTC().Format(time.RFC3339))
	}
}

func (s *Statement) AppendWhereClauseFromTimeCondition() {
	if s.QueryCondition != nil {
		// Append start time
		if s.QueryCondition.StartTime != nil {
			s.AppendWhereClauseWithTime(">=", s.QueryCondition.StartTime.Unix())
		}

		// Append end time
		if s.QueryCondition.EndTime != nil {
			s.AppendWhereClauseWithTime("<=", s.QueryCondition.EndTime.Unix())
		}
	}
}

func (s *Statement) SetOrderClauseFromQueryCondition() {
	switch s.QueryCondition.TimestampOrder {
	case DBCommon.Asc:
		s.OrderClause = "ORDER BY time ASC"
	case DBCommon.Desc:
		s.OrderClause = "ORDER BY time DESC"
	default:
		s.OrderClause = "ORDER BY time ASC"
	}
}

func (s *Statement) SetLimitClauseFromQueryCondition() {
	limit := s.QueryCondition.Limit
	if limit > 0 {
		s.LimitClause = fmt.Sprintf("LIMIT %v", limit)
	}
}

func (s Statement) BuildQueryCmd() string {
	var (
		cmd        = ""
		fieldsStr  = "*"
		groupByStr = ""
	)

	if len(s.SelectedFields) > 0 {
		fieldsStr = ""
		for _, field := range s.SelectedFields {
			fieldsStr += fmt.Sprintf(`"%s",`, field)
		}
		fieldsStr = strings.TrimSuffix(fieldsStr, ",")
	}

	if len(s.GroupByTags) > 0 {
		groupByStr = "GROUP BY "
		for _, field := range s.GroupByTags {
			groupByStr += fmt.Sprintf(`"%s",`, field)
		}
		groupByStr = strings.TrimSuffix(groupByStr, ",")
	}

	cmd = fmt.Sprintf("SELECT %s FROM %s %s %s %s %s",
		fieldsStr, s.Measurement, s.WhereClause,
		groupByStr, s.OrderClause, s.LimitClause)

	return cmd
}
