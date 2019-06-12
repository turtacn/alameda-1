package influxdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/containers-ai/alameda/pkg/utils/log"
	Common "github.com/containers-ai/api/common"
	InfluxDBClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"strings"
	"time"
)

type Database string
type Measurement string

var (
	// ZeroTime is used as a constant of timestamp
	ZeroTime time.Time = time.Unix(0, 0)
)

type InfluxDBEntity struct {
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

type InfluxDBRow struct {
	Name    string
	Tags    map[string]string
	Data    []map[string]string
	Partial bool
}

var (
	scope = log.RegisterScope("influxdb_client", "influxdb client", 0)
)

// InfluxDBRepository interacts with database
type InfluxDBRepository struct {
	Address                string
	Username               string
	Password               string
	RetentionDuration      string
	RetentionShardDuration string
}

// New returns InfluxDBRepository instance
func New(influxCfg *Config) *InfluxDBRepository {
	return &InfluxDBRepository{
		Address:                influxCfg.Address,
		Username:               influxCfg.Username,
		Password:               influxCfg.Password,
		RetentionDuration:      influxCfg.RetentionDuration,
		RetentionShardDuration: influxCfg.RetentionShardDuration,
	}
}

// CreateDatabase creates database
func (influxDBRepository *InfluxDBRepository) CreateDatabase(db string) error {
	_, err := influxDBRepository.QueryDB(fmt.Sprintf("CREATE DATABASE %s", db), db)
	return err
}

// Modify default retention policy
func (influxDBRepository *InfluxDBRepository) ModifyDefaultRetentionPolicy(db string) error {
	duration := influxDBRepository.RetentionDuration
	shardGroupDuration := influxDBRepository.RetentionShardDuration
	retentionCmd := fmt.Sprintf("ALTER RETENTION POLICY \"autogen\" on \"%s\" DURATION %s SHARD DURATION %s", db, duration, shardGroupDuration)
	_, err := influxDBRepository.QueryDB(retentionCmd, db)
	return err
}

func (influxDBRepository *InfluxDBRepository) newHttpClient() InfluxDBClient.Client {
	clnt, err := InfluxDBClient.NewHTTPClient(InfluxDBClient.HTTPConfig{
		Addr:               influxDBRepository.Address,
		Username:           influxDBRepository.Username,
		Password:           influxDBRepository.Password,
		InsecureSkipVerify: true,
	})
	if err != nil {
		scope.Error(err.Error())
	}
	return clnt
}

// WritePoints writes points to database
func (influxDBRepository *InfluxDBRepository) WritePoints(points []*InfluxDBClient.Point, bpCfg InfluxDBClient.BatchPointsConfig) error {
	clnt := influxDBRepository.newHttpClient()
	defer clnt.Close()

	bp, err := InfluxDBClient.NewBatchPoints(bpCfg)
	if err != nil {
		scope.Error(err.Error())
	}

	for _, point := range points {
		bp.AddPoint(point)
	}

	if err := clnt.Write(bp); err != nil {
		if strings.Contains(err.Error(), "database not found") {
			if err = influxDBRepository.CreateDatabase(bpCfg.Database); err != nil {
				scope.Error(err.Error())
				return err
			} else {
				err = influxDBRepository.WritePoints(points, bpCfg)
			}
		}
		if err != nil {
			scope.Error(err.Error())
			fmt.Print(err.Error())
			return err
		}
	}

	return nil
}

// QueryDB queries database
func (influxDBRepository *InfluxDBRepository) QueryDB(cmd, database string) (res []InfluxDBClient.Result, err error) {
	clnt := influxDBRepository.newHttpClient()
	defer clnt.Close()
	q := InfluxDBClient.Query{
		Command:  cmd,
		Database: database,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func PackMap(results []InfluxDBClient.Result) []*InfluxDBRow {
	var rows []*InfluxDBRow

	if len(results[0].Series) == 0 {
		return rows
	}

	for _, result := range results {
		for _, r := range result.Series {
			row := InfluxDBRow{Name: r.Name, Partial: r.Partial}
			row.Tags = r.Tags
			for _, v := range r.Values {
				data := make(map[string]string)
				// Pack tag
				for key, value := range r.Tags {
					data[key] = value
				}
				// Pack values
				for j, col := range r.Columns {
					value := v[j]
					if value != nil {
						switch value.(type) {
						case bool:
							data[col] = strconv.FormatBool(value.(bool))
						case string:
							data[col] = value.(string)
						case json.Number:
							data[col] = value.(json.Number).String()
						case nil:
							data[col] = ""
						default:
							data[col] = value.(string)
						}
					} else {
						data[col] = ""
					}
				}
				row.Data = append(row.Data, data)
			}
			rows = append(rows, &row)
		}
	}

	return NormalizeResult(rows)
}

func NormalizeResult(rows []*InfluxDBRow) []*InfluxDBRow {
	var rowList []*InfluxDBRow

	for _, r := range rows {
		row := InfluxDBRow{Name: r.Name, Partial: r.Partial}
		row.Tags = r.Tags
		for _, d := range r.Data {
			data := make(map[string]string)
			for key, value := range d {
				if strings.HasSuffix(key, "_1") {
					found := false
					newKey := strings.TrimSuffix(key, "_1")
					for k := range data {
						if k == newKey {
							found = true
							if value != "" {
								data[k] = value
							}
							break
						}
					}
					if !found {
						data[key] = value
					}
				} else {
					found := false
					newKey := fmt.Sprintf("%s_1", key)
					for k, v := range data {
						if k == newKey {
							found = true
							if v != "" {
								delete(data, newKey)
								data[key] = v
							} else {
								delete(data, newKey)
								data[key] = value
							}
							break
						}
					}
					if !found {
						data[key] = value
					}
				}
			}
			row.Data = append(row.Data, data)
		}
		rowList = append(rowList, &row)
	}

	return rowList
}

func (influxDBRepository *InfluxDBRepository) AddWhereCondition(whereStr *string, key string, operator string, value string) {
	if value == "" {
		return
	}

	if *whereStr == "" {
		*whereStr += fmt.Sprintf("WHERE \"%s\"%s'%s' ", key, operator, value)
	} else {
		*whereStr += fmt.Sprintf("AND \"%s\"%s'%s' ", key, operator, value)
	}
}

func (influxDBRepository *InfluxDBRepository) AddWhereConditionDirect(whereStr *string, condition string) {
	if condition == "" {
		return
	}

	if *whereStr == "" {
		*whereStr += fmt.Sprintf("WHERE %s ", condition)
	} else {
		*whereStr += fmt.Sprintf("AND %s ", condition)
	}
}

func (influxDBRepository *InfluxDBRepository) AddTimeCondition(whereStr *string, operator string, value int64) {
	if value == 0 {
		return
	}

	tm := time.Unix(int64(value), 0)

	if *whereStr == "" {
		*whereStr += fmt.Sprintf("WHERE time%s'%s' ", operator, tm.UTC().Format(time.RFC3339))
	} else {
		*whereStr += fmt.Sprintf("AND time%s'%s' ", operator, tm.UTC().Format(time.RFC3339))
	}
}

func InfluxResultToReadRawdata(results []InfluxDBClient.Result, query *Common.Query) *Common.ReadRawdata {
	readRawdata := Common.ReadRawdata{Query: query}

	if len(results[0].Series) == 0 {
		return &readRawdata
	}

	for _, result := range results {
		tagsLen := 0
		valuesLen := 0

		// Build columns
		for k := range results[0].Series[0].Tags {
			readRawdata.Columns = append(readRawdata.Columns, string(k))
			tagsLen = tagsLen + 1
		}
		for _, column := range results[0].Series[0].Columns {
			readRawdata.Columns = append(readRawdata.Columns, column)
			valuesLen = valuesLen + 1
		}

		// One series is one group
		for _, row := range result.Series {
			group := Common.Group{}

			// Build values
			for _, value := range row.Values {
				r := Common.Row{}

				// Tags
				for k, v := range readRawdata.Columns {
					r.Values = append(r.Values, row.Tags[v])
					if k >= (tagsLen - 1) {
						break
					}
				}

				// Fields
				for _, v := range value {
					switch v.(type) {
					case bool:
						r.Values = append(r.Values, strconv.FormatBool(v.(bool)))
					case string:
						r.Values = append(r.Values, v.(string))
					case json.Number:
						r.Values = append(r.Values, v.(json.Number).String())
					case nil:
						r.Values = append(r.Values, "")
					default:
						fmt.Println("Error, not support")
						r.Values = append(r.Values, v.(string))
					}
				}
				group.Rows = append(group.Rows, &r)
			}
			readRawdata.Groups = append(readRawdata.Groups, &group)
		}
	}

	return &readRawdata
}

func ReadRawdataToInfluxDBRow(readRawdata *Common.ReadRawdata) []*InfluxDBRow {
	influxDBRows := make([]*InfluxDBRow, 0)

	tagIndex := make([]int, 0)

	// locate tags index
	for _, tag := range readRawdata.GetQuery().GetCondition().GetGroups() {
		for index, column := range readRawdata.GetColumns() {
			if tag == column {
				tagIndex = append(tagIndex, index)
			}
		}
	}

	for _, group := range readRawdata.GetGroups() {
		influxDBRow := InfluxDBRow{
			Name: readRawdata.GetQuery().GetTable(),
			Tags: make(map[string]string),
		}

		for _, row := range group.GetRows() {
			// Pack tags
			for _, v := range tagIndex {
				for _, row := range group.GetRows() {
					influxDBRow.Tags[readRawdata.GetColumns()[v]] = row.GetValues()[v]
				}
			}

			// Pack data
			data := make(map[string]string)
			for index, column := range readRawdata.GetColumns() {
				data[column] = row.GetValues()[index]
			}
			influxDBRow.Data = append(influxDBRow.Data, data)
		}

		influxDBRows = append(influxDBRows, &influxDBRow)
	}

	return influxDBRows
}

func CompareRawdataWithInfluxResults(readRawdata *Common.ReadRawdata, results []InfluxDBClient.Result) error {
	before := PackMap(results)
	after := ReadRawdataToInfluxDBRow(readRawdata)
	message := ""

	for index, row := range after {
		compRow := before[index]

		// Check Name
		if row.Name != compRow.Name {
			message = message + fmt.Sprintf("Name: %s, %s\n", row.Name, compRow.Name)
			fmt.Printf("[ERROR] Name: %s, %s\n", row.Name, compRow.Name)
		}

		// Check Tags
		for key, value := range row.Tags {
			compValue := compRow.Tags[key]
			if compRow.Tags[key] != value {
				message = message + fmt.Sprintf("Tag[%s]: %s, %s\n", key, value, compValue)
				fmt.Printf("[ERROR] Tag[%s]: %s, %s\n", key, value, compValue)
			}
		}

		// Check Data
		for k, v := range row.Data {
			for key, value := range v {
				compValue := compRow.Data[k][key]
				if compValue != value {
					message = message + fmt.Sprintf("Data[%s]: %s, %s\n", key, value, compValue)
					fmt.Printf("[ERROR] Data[%s]: %s, %s\n", key, value, compValue)
				}
			}
		}
	}

	if message != "" {
		return errors.New(message)
	}

	return nil
}
