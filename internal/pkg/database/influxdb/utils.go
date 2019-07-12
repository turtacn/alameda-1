package influxdb

import (
	"encoding/json"
	"fmt"
	Client "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"strings"
	"time"
)

type InfluxEntity struct {
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

type InfluxRow struct {
	Name    string
	Tags    map[string]string
	Data    []map[string]string
	Partial bool
}

func PackMap(results []Client.Result) []*InfluxRow {
	var rows []*InfluxRow

	if len(results[0].Series) == 0 {
		return rows
	}

	for _, result := range results {
		for _, r := range result.Series {
			row := InfluxRow{Name: r.Name, Partial: r.Partial}
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

func NormalizeResult(rows []*InfluxRow) []*InfluxRow {
	var rowList []*InfluxRow

	for _, r := range rows {
		row := InfluxRow{Name: r.Name, Partial: r.Partial}
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
