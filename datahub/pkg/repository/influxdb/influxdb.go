package influxdb

import (
	"time"

	"github.com/containers-ai/alameda/pkg/utils/log"
	client "github.com/influxdata/influxdb/client/v2"
)

type Database string
type Measurement string

type InfluxDBEntity struct {
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

var (
	scope = log.RegisterScope("influxdb_client", "influxdb client", 0)
)

type InfluxDBRepository struct {
	Address  string
	Username string
	Password string
}

func New(influxCfg *Config) *InfluxDBRepository {
	return &InfluxDBRepository{
		Address:  influxCfg.Address,
		Username: influxCfg.Username,
		Password: influxCfg.Password,
	}
}

func (influxDBRepository *InfluxDBRepository) newHttpClient() client.Client {
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
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

func (influxDBRepository *InfluxDBRepository) WritePoints(points []*client.Point, bpCfg client.BatchPointsConfig) {
	clnt := influxDBRepository.newHttpClient()
	defer clnt.Close()

	bp, err := client.NewBatchPoints(bpCfg)
	if err != nil {
		scope.Error(err.Error())
	}

	for _, point := range points {
		bp.AddPoint(point)
	}

	if err := clnt.Write(bp); err != nil {
		scope.Error(err.Error())
	}
}

func (influxDBRepository *InfluxDBRepository) QueryDB(cmd, database string) (res []client.Result, err error) {
	clnt := influxDBRepository.newHttpClient()
	defer clnt.Close()
	q := client.Query{
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
