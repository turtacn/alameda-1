package influxdb

import (
	"errors"
	"fmt"
	Client "github.com/influxdata/influxdb/client/v2"
	"strings"
	"time"
)

// Create database
func (p *InfluxClient) CreateDatabase(db string, retry int) error {
	counter := 0
	err := errors.New("")

	for true {
		counter += 1
		_, err = p.QueryDB(fmt.Sprintf("CREATE DATABASE %s", db), db)
		if err == nil {
			break
		}
		scope.Error(err.Error())
		scope.Errorf("failed to create database(%s), try again", db)
		if retry < 0 || (retry > 0 && retry == counter) {
			break
		}
		time.Sleep(3 * time.Second)
	}

	return err
}

// Delete database
func (p *InfluxClient) DeleteDatabase(db string) error {
	_, err := p.QueryDB(fmt.Sprintf("DROP DATABASE %s", db), db)
	return err
}

// Delete measurement
func (p *InfluxClient) DeleteMeasurement(db, measurement string) error {
	_, err := p.QueryDB(fmt.Sprintf("DROP MEASUREMENT %s", measurement), db)
	return err
}

func (p *InfluxClient) MeasurementExist(db, measurement string) bool {
	if response, err := p.QueryDB(fmt.Sprintf("SHOW FIELD KEYS FROM %s", measurement), db); err == nil {
		if len(response) > 0 && response[0].Series != nil {
			return true
		}
	}
	return false
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
			if err = p.CreateDatabase(bpCfg.Database, -1); err != nil {
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
func (p *InfluxClient) ModifyDefaultRetentionPolicy(db string, retry int) error {
	duration := p.RetentionDuration
	shardGroupDuration := p.RetentionShardDuration
	counter := 0
	err := errors.New("")
	retentionCmd := fmt.Sprintf("ALTER RETENTION POLICY \"autogen\" on \"%s\" DURATION %s SHARD DURATION %s", db, duration, shardGroupDuration)
	for true {
		counter += 1
		_, err = p.QueryDB(retentionCmd, db)
		if err == nil {
			break
		}
		scope.Error(err.Error())
		scope.Errorf("failed to modify retention policy on database(%s), try again", db)
		if retry < 0 || (retry > 0 && retry == counter) {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return err
}

func (p *InfluxClient) Ping() error {
	client := p.newHttpClient()
	defer client.Close()

	duration, version, err := client.Ping(10 * time.Second)
	if err != nil {
		scope.Error("failed to ping to InfluxDB")
		return err
	}

	scope.Info(duration.String())
	scope.Info(version)

	return nil
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
