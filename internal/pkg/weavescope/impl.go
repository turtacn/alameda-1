package weavescope

import (
	"fmt"
	ApiWeavescope "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/weavescope"
	"io/ioutil"
	"net/http"
)

func (w *WeaveScopeClient) ListWeaveScopeHosts(in *ApiWeavescope.ListWeaveScopeHostsRequest) (string, error) {
	url := fmt.Sprintf("%s%s", w.URL, "/api/topology/hosts")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) GetWeaveScopeHostDetails(in *ApiWeavescope.ListWeaveScopeHostsRequest) (string, error) {
	url := fmt.Sprintf("%s%s%s;<host>", w.URL, "/api/topology/hosts/", in.GetHostId())

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) ListWeaveScopePods(in *ApiWeavescope.ListWeaveScopePodsRequest) (string, error) {
	url := fmt.Sprintf("%s%s", w.URL, "/api/topology/pods")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) GetWeaveScopePodDetails(in *ApiWeavescope.ListWeaveScopePodsRequest) (string, error) {
	url := fmt.Sprintf("%s%s%s;<pod>", w.URL, "/api/topology/pods/", in.GetPodId())

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) ListWeaveScopeContainers(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	url := fmt.Sprintf("%s%s", w.URL, "/api/topology/containers")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) ListWeaveScopeContainersByHostname(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	url := fmt.Sprintf("%s%s", w.URL, "/api/topology/containers-by-hostname")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) ListWeaveScopeContainersByImage(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	url := fmt.Sprintf("%s%s", w.URL, "/api/topology/containers-by-image")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}

func (w *WeaveScopeClient) GetWeaveScopeContainerDetails(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	url := fmt.Sprintf("%s%s%s;<container>", w.URL, "/api/topology/containers/", in.GetContainerId())

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	readBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(readBody), nil
}
