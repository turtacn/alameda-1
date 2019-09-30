package nodeinfo

import (
	"testing"

	"github.com/ghodss/yaml"
	"io/ioutil"
	"k8s.io/api/core/v1"
)

func TestNewNodeInfo(t *testing.T) {
	var nodeList v1.NodeList
	dat, err := ioutil.ReadFile("nodeinfo_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	err = yaml.Unmarshal(dat, &nodeList)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodeList.Items {
		nodeInfo, err := NewNodeInfo(n)
		if err != nil {
			t.Fatal(err)
			continue
		}
		t.Logf("%+v", nodeInfo)
	}
}
