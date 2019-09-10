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
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

func EventHTMLMsg(evt *datahub_v1alpha1.Event) string {
	levelMap := viper.GetStringMap("eventLevel")
	eventMap := viper.GetStringMap("eventType")
	return fmt.Sprintf(`
	<html>
		<body>
			Federator.ai Event Notification<br>
			###############################################################<br>
			<table cellspacing="5" cellpadding="0">
				<tr><td align="left">Cluster Id:</td><td>%s</td></tr>
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
	`, evt.GetClusterId(), time.Unix(evt.Time.GetSeconds(), 0).Format(time.RFC3339),
		strings.Title(levelMap[strconv.FormatInt(int64(evt.Level), 10)].(string)), evt.Message,
		eventMap[strconv.FormatInt(int64(evt.Type), 10)].(string), evt.Subject.Name,
		evt.Subject.Kind, evt.Subject.Namespace)
}

func GetClusterUID(k8sClient client.Client) (string, error) {
	possibleNSList := []string{
		"kube-service-catalog", "kube-public",
	}
	var errorList []string
	clusterId := ""
	for _, possibleNS := range possibleNSList {
		clusterInfoCM := &corev1.ConfigMap{}
		err := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "cluster-info",
			Namespace: possibleNS,
		}, clusterInfoCM)
		if err == nil {
			return string(clusterInfoCM.GetUID()), nil
		} else if !k8sErrors.IsNotFound(err) {
			errorList = append(errorList, err.Error())
		}
	}
	if len(errorList) == 0 {
		return clusterId, fmt.Errorf("no cluster info found")
	}
	return clusterId, errors.New(strings.Join(errorList, ","))
}
