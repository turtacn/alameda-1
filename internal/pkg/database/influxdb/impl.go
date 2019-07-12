package influxdb

import (
	"fmt"
	Client "github.com/influxdata/influxdb/client/v2"
	"strings"
)

// Creates database
func (p *InfluxClient) CreateDatabase(db string) error {
	_, err := p.QueryDB(fmt.Sprintf("CREATE DATABASE %s", db), db)
	return err
}

// Write points to database
func (p *InfluxClient) WritePoints(points []*Client.Point, bpCfg Client.BatchPointsConfig) error {
	client := p.newHttpClient()
	defer client.Close()

	bp, err := Client.NewBatchPoints(bpCfg)
	if err != nil {
		scope.Error(err.Error())
	}

	for _, point := range points {
		bp.AddPoint(point)
	}

	if err := client.Write(bp); err != nil {
		if strings.Contains(err.Error(), "database not found") {
			if err = p.CreateDatabase(bpCfg.Database); err != nil {
				scope.Error(err.Error())
				return err
			} else {
				err = p.WritePoints(points, bpCfg)
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

// Query database
func (p *InfluxClient) QueryDB(cmd, database string) (res []Client.Result, err error) {
	client := p.newHttpClient()
	defer client.Close()

	q := Client.Query{
		Command:  cmd,
		Database: database,
	}

	if response, err := client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}

	return res, nil
}

// Modify default retention policy
func (p *InfluxClient) ModifyDefaultRetentionPolicy(db string) error {
	duration := p.RetentionDuration
	shardGroupDuration := p.RetentionShardDuration
	retentionCmd := fmt.Sprintf("ALTER RETENTION POLICY \"autogen\" on \"%s\" DURATION %s SHARD DURATION %s", db, duration, shardGroupDuration)
	_, err := p.QueryDB(retentionCmd, db)
	return err
}

func (p *InfluxClient) newHttpClient() Client.Client {
	client, err := Client.NewHTTPClient(Client.HTTPConfig{
		Addr:               p.Address,
		Username:           p.Username,
		Password:           p.Password,
		InsecureSkipVerify: true,
	})
	if err != nil {
		scope.Error(err.Error())
	}
	return client
}
