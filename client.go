package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	DeployNamespace  = os.Getenv("DEPLOY_NAMESPACE")
	ImageNamespace   = os.Getenv("IMAGE_NAMESPACE")
	OpenshiftApiHost = os.Getenv("OPENSHIFT_API_HOST")
)

const (
	JobPrefix = "concord-task-"
	TokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

var (
	ImageNotFoundError = errors.New("no tags exist for the provided image stream")
	JobNotFoundError   = errors.New("no job was found with the provided name")
)

type Client interface {
	CancelTag(string) error
	GetImageTag(string, string) (string, error)
	StartTag(string, string, json.RawMessage) error
}

type OpenshiftClient struct {
	BearerToken string
	Client      *http.Client
}

type ApiResponse struct {
	Code   int               `json:"code"`
	Status ImageStreamStatus `json:"status"`
}

type ImageStreamStatus struct {
	Tags []ImageStreamTags `json:"tags"`
}

type ImageStreamTags struct {
	Tag   string                   `json:"tag"`
	Items []map[string]interface{} `json:"items"`
}

func NewOpenshiftClient() (*OpenshiftClient, error) {
	tokenFd, err := os.Open(TokenPath)
	if err != nil {
		return nil, err
	}
	token, err := ioutil.ReadAll(tokenFd)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return &OpenshiftClient{BearerToken: string(token), Client: client}, nil
}

func (oc *OpenshiftClient) CancelTask(name string) error {
	url := fmt.Sprintf(`https://%s:8443/apis/batch/v1/namespaces/%s/jobs/%s`,
		OpenshiftApiHost, DeployNamespace, fmt.Sprintf("%s%s", JobPrefix, name))
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oc.BearerToken))
	resp, err := oc.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response ApiResponse
	json.Unmarshal(body, &response)
	if response.Code == 404 {
		return JobNotFoundError
	}
	return nil
}

func (oc *OpenshiftClient) GetImageTag(name, tag string) (string, error) {
	url := fmt.Sprintf(`https://%s:8443/oapi/v1/namespaces/%s/imagestreams/%s`,
		OpenshiftApiHost, ImageNamespace, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oc.BearerToken))
	resp, err := oc.Client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var response ApiResponse
	json.Unmarshal(body, &response)
	for i, is := range response.Status.Tags {
		if is.Tag == tag && len(response.Status.Tags[i].Items) > 0 {
			return response.Status.Tags[i].Items[0]["dockerImageReference"].(string), nil
		}
	}
	return "", ImageNotFoundError
}

func (oc *OpenshiftClient) StartTask(name, namespace string, spec) error {
	obj := `{
        "apiVersion": "batch/v1",
        "kind": "Job",
        "metadata": {
            "name": %s%s
        },
        "spec": {
            "containers": [
                {
                    "name": %s%s,
                    "image": %s,
                }
            ],
            "restartPolicy" "Never"
        }
    }`
}
