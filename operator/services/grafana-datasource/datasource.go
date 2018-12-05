package grafanadatasource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scope = logUtil.RegisterScope("grafanadatasource", "grafana datasource log", 0)
)

type PredictDeployment struct {
	Namespace      string
	Name           string
	DeploymentName string
	DeploymentUID  string
}

type PredictPod struct {
	Namespace string
	Name      string
	PodName   string
	PodUID    string
}
type PredictContainer struct {
	Namespace      string
	Name           string
	DeploymentName string
	PodName        string
	PodUID         string
	ContainerName  string
	RawPredict     map[autoscalingv1alpha1.ResourceType]autoscalingv1alpha1.TimeSeriesData
}

type GrafanaDataSource struct {
	Manager   manager.Manager
	K8SClient client.Client
}

func NewGrafanaDataSource(mgr manager.Manager, bindPort uint16) *GrafanaDataSource {
	gs := GrafanaDataSource{
		Manager:   mgr,
		K8SClient: mgr.GetClient(),
	}
	gs.initHttp(bindPort)
	return &gs
}

func (gds *GrafanaDataSource) initHttp(bindPort uint16) {
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		scope.Info("/search")
		gds.handleSearch(w, r)
	})
	http.HandleFunc("/annotations", func(w http.ResponseWriter, r *http.Request) {
		scope.Info("/annotation")
		fmt.Fprintf(w, "")
	})
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		scope.Info("/query")
		gds.handleQuery(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		scope.Info("/")
		fmt.Fprintf(w, "")
	})

	scope.Infof(fmt.Sprintf("Grafana binding port is %s", strconv.Itoa(int(bindPort))))
	scope.Error(http.ListenAndServe(":"+strconv.Itoa(int(bindPort)), nil).Error())
}

func (gds *GrafanaDataSource) handleQuery(w http.ResponseWriter, r *http.Request) {
	queryRequest := &QueryRequest{}
	err := json.NewDecoder(r.Body).Decode(queryRequest)
	if err != nil {
		scope.Error(err.Error())
		fmt.Fprintf(w, "")
	} else {
		if len(queryRequest.Targets) > 0 {
			targetType := queryRequest.Targets[0].Type
			target := queryRequest.Targets[0].Target
			scope.Infof(fmt.Sprintf("Target Type: %s, Target: %s", targetType, target))

			if nnStrs := strings.Split(target, "\\/"); targetType == "table" && len(nnStrs) == 2 {
				namespace := nnStrs[0]
				name := nnStrs[1]
				scope.Info(fmt.Sprintf("List prediction containers for visualization (%s/%s)", namespace, name))
				resList := []QueryResponse{}
				rows := [][]interface{}{}
				pdContainers := gds.getPredictionContainers(namespace, name)
				if pdContainers != nil {
					for _, pdContainer := range pdContainers {
						rows = append(rows, []interface{}{pdContainer.Namespace, pdContainer.Name, pdContainer.DeploymentName, pdContainer.PodName, pdContainer.ContainerName})
					}
				}
				resList = append(resList, QueryResponse{
					Columns: []SearchResponse{
						SearchResponse{
							Text: "Namesapce", Type: "string",
						}, SearchResponse{
							Text: "Name", Type: "string",
						}, SearchResponse{
							Text: "Deployment", Type: "string",
						}, SearchResponse{
							Text: "Pod", Type: "string",
						}, SearchResponse{
							Text: "Container", Type: "string",
						},
					},
					Rows: rows,
					Type: "table",
				})
				resListBin, _ := json.MarshalIndent(resList, "", "  ")
				fmt.Fprintf(w, string(resListBin))
			} else if resNNIDStrs := strings.Split(target, ":"); targetType == "timeserie" && len(resNNIDStrs) == 3 {
				if nnStrs := strings.Split(resNNIDStrs[1], "\\/"); len(nnStrs) == 2 {
					namespace := nnStrs[0]
					name := nnStrs[1]
					resList := []TimeSerie{}
					if pdContainers := gds.getPredictionContainers(namespace, name); pdContainers != nil {
						nowSec := time.Now().UnixNano() / 1000000000
						for _, pdContainer := range pdContainers {
							dataPoints := [][]float64{}
							if resNNStrs := strings.Split(target, ":"); len(resNNStrs) == 3 {
								podUID := resNNStrs[2]
								if podUID != pdContainer.PodUID {
									continue
								}
								for resource, tsData := range pdContainer.RawPredict {
									for _, predictData := range tsData.PredictData {
										value, valueErr := strconv.ParseFloat(predictData.Value, 64)
										time := float64(predictData.Time)
										if valueErr == nil && strings.ToLower(resNNStrs[0]) == strings.ToLower(string(resource)) && float64(nowSec) < time {
											dataPoints = append(dataPoints, []float64{value, time * 1000})
										}
									}
								}
							}
							resList = append(resList, TimeSerie{
								Target:     pdContainer.ContainerName, //target,
								DataPoints: dataPoints,
							})
						}
						resListBin, _ := json.MarshalIndent(resList, "", "  ")
						fmt.Fprintf(w, string(resListBin))
					}
				}
			}

		}
		bin, _ := json.MarshalIndent(*queryRequest, "", "  ")
		scope.Info(string(bin))
	}
}
func (gds *GrafanaDataSource) handleSearch(w http.ResponseWriter, r *http.Request) {
	searchRequest := &SearchRequest{}
	err := json.NewDecoder(r.Body).Decode(searchRequest)
	if err != nil {
		scope.Error(err.Error())
		fmt.Fprintf(w, "")
	} else {
		resList := []SearchResponse{}
		scope.Infof(fmt.Sprintf("The target search request is %s", searchRequest.Target))
		if searchRequest.Target == "alamedaresourceprediction" {
			if predictions := gds.listAlamedaPrediction(); predictions != nil {
				for _, prediction := range predictions {
					resList = append(resList, SearchResponse{
						Type:  "alamedaresourceprediction",
						Text:  prediction.GetNamespace() + "/" + prediction.GetName(),
						Value: prediction.GetNamespace() + "/" + prediction.GetName(),
					})
				}
			}
		} else if strings.HasPrefix(searchRequest.Target, "predictioncontainer:") {
			spStrs := strings.Split(searchRequest.Target, "predictioncontainer:")
			if len(spStrs) == 2 {
				nnStrs := strings.Split(spStrs[1], "\\/")
				if len(nnStrs) == 2 {
					namespace := nnStrs[0]
					name := nnStrs[1]
					pdContainers := gds.getPredictionContainers(namespace, name)
					if pdContainers != nil {
						for _, pdContainer := range pdContainers {
							resList = append(resList, SearchResponse{
								Type:  "predictioncontainer",
								Text:  pdContainer.PodName + "/" + pdContainer.ContainerName,
								Value: pdContainer.Namespace + "/" + pdContainer.Name + "/" + pdContainer.PodName + "/" + pdContainer.ContainerName,
							})

						}
					}
				}
			}
		} else if strings.HasPrefix(searchRequest.Target, "predictiondeployment:") {
			spStrs := strings.Split(searchRequest.Target, "predictiondeployment:")
			if len(spStrs) == 2 {
				nnStrs := strings.Split(spStrs[1], "\\/")
				if len(nnStrs) == 2 {
					namespace := nnStrs[0]
					name := nnStrs[1]
					pdDeployments := gds.getPredictionDeployments(namespace, name)
					if pdDeployments != nil {
						for _, pdDeployment := range pdDeployments {
							resList = append(resList, SearchResponse{
								Type:  "predictiondeployment",
								Text:  pdDeployment.DeploymentName,
								Value: pdDeployment.DeploymentUID,
							})
						}
					}
				}
			}
		} else if strings.HasPrefix(searchRequest.Target, "predictpod:") {
			if spStrs := strings.Split(searchRequest.Target, "predictpod:"); len(spStrs) == 2 {
				if nnStrs := strings.Split(spStrs[1], "\\/"); len(nnStrs) == 2 {
					namespace := nnStrs[0]
					name := nnStrs[1]
					if predictpods := gds.getPredictionPods(namespace, name); predictpods != nil {
						for _, predictpod := range predictpods {
							resList = append(resList, SearchResponse{
								Type:  "predictpod",
								Text:  predictpod.PodName,
								Value: predictpod.PodUID,
							})
						}
					}
				}
			}
		}

		resListBin, _ := json.MarshalIndent(resList, "", "  ")
		fmt.Fprintf(w, string(resListBin))
	}
}

func (gds *GrafanaDataSource) listAlamedaPrediction() []autoscalingv1alpha1.AlamedaResourcePrediction {

	alamedaResourcePredictionList := &autoscalingv1alpha1.AlamedaResourcePredictionList{}
	err := gds.K8SClient.List(context.TODO(),
		client.InNamespace(""),
		alamedaResourcePredictionList)
	if err != nil {
		scope.Error(err.Error())
	} else {
		listBin, err := json.MarshalIndent(*alamedaResourcePredictionList, "", "  ")
		if err != nil {
			scope.Error(err.Error())
		} else {
			scope.Info(string(listBin))
			return alamedaResourcePredictionList.Items
		}
	}
	return nil
}
func (gds *GrafanaDataSource) getAlamedaPrediction(namespace, name string) *autoscalingv1alpha1.AlamedaResourcePrediction {

	alamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{}
	err := gds.K8SClient.Get(context.TODO(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
		alamedaResourcePrediction)
	if err != nil {
		scope.Error(err.Error())
	} else {
		predictBin, err := json.MarshalIndent(*alamedaResourcePrediction, "", "  ")
		if err != nil {
			scope.Error(err.Error())
		} else {
			scope.Info(string(predictBin))
		}
		return alamedaResourcePrediction
	}
	return nil
}

func (gds *GrafanaDataSource) getPredictionDeployments(namespace, name string) []PredictDeployment {
	PredictDeployments := []PredictDeployment{}
	predict := gds.getAlamedaPrediction(namespace, name)
	if predict == nil {
		return nil
	}
	for _, deploy := range predict.Status.Prediction.Deployments {
		deploymentName := deploy.Name
		deploymentUID := deploy.UID

		scope.Infof(fmt.Sprintf("Namespace: %s,Name: %s,Deployment name: %s, Deployment UID: %s", namespace, name, deploymentName, deploymentUID))
		PredictDeployments = append(PredictDeployments, PredictDeployment{
			Namespace:      namespace,
			Name:           name,
			DeploymentName: deploymentName,
			DeploymentUID:  deploymentUID,
		})
	}
	return PredictDeployments
}

func (gds *GrafanaDataSource) getPredictionPods(namespace, name string) []PredictPod {
	predictPods := []PredictPod{}
	predict := gds.getAlamedaPrediction(namespace, name)
	if predict == nil {
		return nil
	}
	for _, deploy := range predict.Status.Prediction.Deployments {
		for podUID, pod := range deploy.Pods {
			podName := pod.Name
			scope.Infof(fmt.Sprintf("Namespace: %s,Name: %s,Pod name: %s, Pod UID: %s", namespace, name, podName, podUID))
			predictPods = append(predictPods, PredictPod{
				Namespace: namespace,
				Name:      name,
				PodName:   podName,
				PodUID:    string(podUID),
			})
		}
	}
	return predictPods
}

func (gds *GrafanaDataSource) getPredictionContainers(namespace, name string) []PredictContainer {
	predictContainers := []PredictContainer{}
	predict := gds.getAlamedaPrediction(namespace, name)
	if predict == nil {
		return nil
	}
	for _, deploy := range predict.Status.Prediction.Deployments {
		for podUID, pod := range deploy.Pods {
			for _, container := range pod.Containers {
				podName := pod.Name
				containerName := container.Name
				scope.Infof(fmt.Sprintf("Namespace: %s,Name: %s,Pod name: %s, Container Name: %s", namespace, name, podName, containerName))
				predictContainers = append(predictContainers, PredictContainer{
					Namespace:      namespace,
					Name:           name,
					DeploymentName: deploy.Name,
					PodName:        podName,
					PodUID:         string(podUID),
					ContainerName:  containerName,
					RawPredict:     container.RawPredict,
				})
			}
		}
	}
	return predictContainers
}
