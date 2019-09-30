package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
)

const (
	ec2LinkLocalAddress                  = "http://169.254.169.254"
	ec2InstanceIdentityDocumentsEndpoint = "/latest/dynamic/instance-identity/document"
)

func OnGCE() bool {
	return metadata.OnGCE()
}

func OnEC2() bool {

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Get(fmt.Sprintf("%s%s", ec2LinkLocalAddress, ec2InstanceIdentityDocumentsEndpoint))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return false
	}
	if _, exist := data["region"]; exist {
		return true
	}

	return false
}

func GetEC2Region() string {

	region := ""

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Get(fmt.Sprintf("%s%s", ec2LinkLocalAddress, ec2InstanceIdentityDocumentsEndpoint))
	if err != nil {
		return region
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return region
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return region
	}
	if v, exist := data["region"]; exist {
		if r, ok := v.(string); ok {
			return r
		}
	}

	return region
}
