package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
	Corev1 "k8s.io/api/core/v1"
	K8SErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

)

type NodeInfo struct {
	UID           string
	Name          string
	RoleMaster    bool
	RoleCompute   bool
	RoleInfra     bool
	Hostname      string
	IPInternal    []string
	IPExternal    []string
}

type ClusterInfo struct {
	UID                string
	Nodes              []NodeInfo
	MasterNodeHostname string
	MasterNodeIP       string
	MasterNodeUID      string
}

func RemoveEmptyStr(strList []string) []string {
	ret := []string{}
	for _, str := range strList {
		if str != "" {
			ret = append(ret, str)
		}
	}
	return ret
}

func EventEmailSubject(evt *datahub_v1alpha1.Event) string {
	msg := evt.GetMessage()
	levelMap := viper.GetStringMap("eventLevel")
	level := evt.GetLevel()
	return fmt.Sprintf("Federator.ai Notification: %s - %s",
		strings.Title(levelMap[strconv.FormatInt(int64(level), 10)].(string)), msg)
}

func EventHTMLMsg(evt *datahub_v1alpha1.Event, clusterInfo *ClusterInfo) string {
	var cInfo *ClusterInfo
	levelMap := viper.GetStringMap("eventLevel")
	eventMap := viper.GetStringMap("eventType")
	evtClusterId := evt.GetClusterId()
	cInfoUnknown := ClusterInfo{UID: evtClusterId}
	if evtClusterId == clusterInfo.UID {
		cInfo = clusterInfo
	} else {
		cInfo = &cInfoUnknown
	}
	return fmt.Sprintf(`
	<html>
		<body>
			Federator.ai Event Notification<br>
			###############################################################<br>
			<table cellspacing="5" cellpadding="0">
				<tr><td align="left">Cluster Id:</td><td>%s</td></tr>
				<tr><td align="left">Master Node Hostname:</td><td>%s</td></tr>
				<tr><td align="left">Master Node IP:</td><td>%s</td></tr>
				<tr><td align="left">Time:</td><td>%s</td></tr>
				<tr><td align="left">Level:</td><td>%s</td></tr>
				<tr><td align="left">Message:</td><td>%s</td></tr>
				<tr><td align="left">Event Type:</td><td>%s</td></tr>
				<tr><td align="left">Resource Name:</td><td>%s</td></tr>
				<tr><td align="left">Resource Kind:</td><td>%s</td></tr>
				<tr><td align="left">Namespace:</td><td>%s</td></tr>
			</table>
		</body>
	</html>
	`, cInfo.UID, cInfo.MasterNodeHostname, cInfo.MasterNodeIP,
		time.Unix(evt.Time.GetSeconds(), 0).Format(time.RFC3339),
		strings.Title(levelMap[strconv.FormatInt(int64(evt.Level), 10)].(string)), evt.Message,
		eventMap[strconv.FormatInt(int64(evt.Type), 10)].(string), evt.Subject.Name,
		evt.Subject.Kind, evt.Subject.Namespace)
}

func GetClusterInfo(k8sClient client.Client) (ClusterInfo, error) {
	nodeList := &Corev1.NodeList{}
	errorList := make([]string, 0)
	cInfo := ClusterInfo{}
	nodeFound := false

	err := k8sClient.List(context.Background(), nodeList)
	if err == nil {
		for _, node := range nodeList.Items {
			isMasterNode := false
			newNode := NodeInfo{}
			newNode.Name = node.ObjectMeta.Name
			newNode.UID = string(node.ObjectMeta.GetUID())
			for label, _ := range node.ObjectMeta.Labels {
				switch label {
				case "node-role.kubernetes.io/master":
					newNode.RoleMaster = true
					if isMasterNode == false {
						isMasterNode = true
					}
				case "node-role.kubernetes.io/infra":
					newNode.RoleInfra = true
				case "node-role.kubernetes.io/compute":
					newNode.RoleCompute = true
				}
			}
			addresses := node.Status.Addresses
			for _, addr := range addresses {
				switch addr.Type {
				case Corev1.NodeHostName:
					newNode.Hostname = addr.Address
				case Corev1.NodeInternalIP:
					newNode.IPInternal = append(newNode.IPInternal, addr.Address)
				case Corev1.NodeExternalIP:
					newNode.IPExternal = append(newNode.IPExternal, addr.Address)
				}
			}
			if isMasterNode == true && nodeFound == false {
				cInfo.MasterNodeHostname = newNode.Hostname
				cInfo.MasterNodeUID = newNode.UID
				if cInfo.MasterNodeIP == "" {
					if len(newNode.IPExternal) > 0 {
						cInfo.MasterNodeIP = newNode.IPExternal[0]
					} else if len(newNode.IPInternal) > 0 {
						cInfo.MasterNodeIP = newNode.IPInternal[0]
					}
				}
				nodeFound = true
			}
			cInfo.Nodes = append(cInfo.Nodes, newNode)
		}
		return cInfo, nil
	} else if !K8SErrors.IsNotFound(err) {
		errorList = append(errorList, err.Error())
	}

	if len(errorList) == 0 {
		return cInfo, fmt.Errorf("no nodeList info found")
	}

	return cInfo, errors.New(strings.Join(errorList, ","))
}