package influxdb

import (
	"fmt"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	Common "github.com/containers-ai/api/common"
	"strings"
	"time"
)

type FunctionType int

const (
	NoneFunction FunctionType = iota
	Aggregate
	Select
)

type Statement struct {
	QueryCondition *DBCommon.QueryCondition
	Database       Database
	Measurement    Measurement
	SelectedFields []string
	GroupByTags    []string
	Function       *Function
	WhereClause    string
	TimeClause     string
	OrderClause    string
	LimitClause    string
}

type Function struct {
	FuncType FunctionType
	FuncName string
	Target   string
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
		Function:       nil,
		WhereClause:    query.GetCondition().GetWhereClause(),
	}

	return &statement
}

func (s *Statement) AppendWhereClause(operator, key, op, value string) {
	if value == "" {
		return
	}

	if s.WhereClause == "" {
		s.WhereClause += fmt.Sprintf("WHERE \"%s\"%s'%s' ", key, op, value)
	} else {
		s.WhereClause += fmt.Sprintf("%s \"%s\"%s'%s' ", operator, key, op, value)
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

func (s *Statement) AppendWhereClauseDirectly(operator, condition string) {
	if condition == "" {
		return
	}

	if s.WhereClause == "" {
		s.WhereClause += fmt.Sprintf("WHERE %s ", condition)
	} else {
		s.WhereClause += fmt.Sprintf("%s %s ", operator, condition)
	}
}

func (s *Statement) AppendWhereClauseWithTime(operator string, value int64) {
	if value == 0 {
		return
	}

	tm := time.Unix(int64(value), 0)

	if s.WhereClause == "" && s.TimeClause == "" {
		s.TimeClause += fmt.Sprintf("WHERE time%s'%s' ", operator, tm.UTC().Format(time.RFC3339))
	} else {
		s.TimeClause += fmt.Sprintf("AND time%s'%s' ", operator, tm.UTC().Format(time.RFC3339))
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

func (s *Statement) SetFunction(funcType FunctionType, funcName, target string) {
	s.Function = &Function{}
	s.Function.FuncType = funcType
	s.Function.FuncName = funcName
	s.Function.Target = target
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

func (s *Statement) GenerateCondition(keyList, valueList []string, op string) string {
	condition := ""
	for i := 0; i < len(keyList); i++ {
		if valueList[i] != "" {
			condition += fmt.Sprintf("\"%s\"='%s' %s ", keyList[i], valueList[i], op)
		}
	}
	condition = strings.TrimSuffix(condition, fmt.Sprintf("%s ", op))
	if condition != "" {
		condition = "(" + condition + ")"
	}
	return condition
}

func (s *Statement) GenerateConditionByList(conditionList []string, op string) string {
	condition := ""

	for _, conditionStr := range conditionList {
		condition += fmt.Sprintf("%s %s ", conditionStr, op)
	}

	condition = strings.TrimSuffix(condition, fmt.Sprintf("%s ", op))
	if condition != "" {
		condition = "(" + condition + ")"
	}
	return condition
}

func (s *Statement) BuildQueryCmd() string {
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

	if s.Function != nil {
		switch s.Function.FuncType {
		case Aggregate:
			fieldsStr = fmt.Sprintf("%s(%s)", s.Function.FuncName, fieldsStr)
		case Select:
			fieldsStr = fmt.Sprintf("%s(%s)", s.Function.FuncName, fieldsStr)
		}
		if s.Function.Target != "" {
			fieldsStr = fmt.Sprintf("%s as %s", fieldsStr, s.Function.Target)
		}
	}

	if len(s.GroupByTags) > 0 {
		groupByStr = "GROUP BY "
		for _, field := range s.GroupByTags {
			if strings.HasPrefix(field, "time(") {
				groupByStr += field
			} else {
				groupByStr += fmt.Sprintf(`"%s",`, field)
			}
		}
		groupByStr = strings.TrimSuffix(groupByStr, ",")
	}

	if s.WhereClause != "" {
		index := strings.Index(s.WhereClause, "WHERE") + 6

		tempClause := s.WhereClause
		tempClause = tempClause + ")"
		tempClause = tempClause[:index] + "(" + tempClause[index:]
		s.WhereClause = tempClause
	}

	cmd = fmt.Sprintf("SELECT %s FROM \"%s\" %s %s %s %s %s",
		fieldsStr, s.Measurement, s.WhereClause, s.TimeClause,
		groupByStr, s.OrderClause, s.LimitClause)

	return cmd
}

func (s *Statement) BuildDropCmd() string {
	cmd := ""

	if s.WhereClause != "" {
		index := strings.Index(s.WhereClause, "WHERE") + 6

		tempClause := s.WhereClause
		tempClause = tempClause + ")"
		tempClause = tempClause[:index] + "(" + tempClause[index:]
		s.WhereClause = tempClause
	}

	cmd = fmt.Sprintf("DROP SERIES FROM \"%s\" %s %s", s.Measurement, s.WhereClause, s.TimeClause)

	return cmd
}

func (s *Statement) Clear() {

}
