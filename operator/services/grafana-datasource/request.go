package grafanadatasource

// grafana simplejson 1.4.0
type SearchRequest struct {
	Target string `json:"target"`
}

type QueryRequest struct {
	PanelId       uint          `json:"panelId"`
	Interval      string        `json:"interval"`
	IntervalMs    uint          `json:"intervalMs"`
	Format        string        `json:"format"`
	MaxDataPoints uint          `json:"maxDataPoints"`
	AdhocFilters  []AdhocFilter `json:"adhocFilters"`
	RangeRaw      RangeRaw      `json:"rangeRaw"`
	Targets       []Target      `json:"targets"`
}

type AdhocFilter struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Range struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Raw  RangeRaw `json:"raw"`
}
type RangeRaw struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Target struct {
	Target string `json:"target"`
	RefId  string `json:"refId"`
	Type   string `json:"type"`
}

type SearchResponse struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Value string `json:"value"`
}

type QueryResponse struct {
	Columns []SearchResponse `json:"columns"`
	Rows    [][]interface{}  `json:"rows"`
	Type    string           `json:"type"`
}

type TimeSerie struct {
	Target     string      `json:"target"`
	DataPoints [][]float64 `json:"datapoints"`
}
