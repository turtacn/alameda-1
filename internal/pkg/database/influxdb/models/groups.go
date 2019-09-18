package models

import (
	"encoding/json"
	Models "github.com/influxdata/influxdb/models"
	"strconv"
)

type InfluxGroup struct {
	Models.Row
}

func (g *InfluxGroup) GetRowNum() int {
	return len(g.Values)
}

func (g *InfluxGroup) GetRow(index int) map[string]string {
	data := make(map[string]string)

	// Pack tag
	for key, value := range g.Tags {
		data[key] = value
	}

	// Pack values
	values := g.Values[index]
	for j, col := range g.Columns {
		value := values[j]
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

	return data
}

func (g *InfluxGroup) GetRows() []map[string]string {
	data := make([]map[string]string, 0)

	for i := 0; i< len(g.Values); i ++ {
		data = append(data, g.GetRow(i))
	}

	return data
}
