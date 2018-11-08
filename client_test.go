package main

import (
	"testing"
)

func TestOpenshiftClientGetImageTag(t *testing.T) {
	client, err := NewOpenshiftClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetImageTag("agilis-pyats-runner", "latest"); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetImageTag("test", "latest"); err != ImageNotFoundError {
		t.Fatal(err)
	}
}

func TestOpenshiftClientCancelTask(t *testing.T) {
	var table = []struct {
		Name   string
		Create bool
		Err    error
	}{
		{
			"hello-world",
			false,
			JobNotFoundError,
		},
	}

	for _, tt := range table {
		client, err := NewOpenshiftClient()
		if err != nil {
			t.Fatal(err)
		}
		if err := client.CancelTask(tt.Name); err != tt.Err {
			t.Fatal(err)
		}
	}
}

func TestOpenshiftClientStartTask(t *testing.T) {
	client, err := NewOpenshiftClient()
	if err != nil {
		t.Fatal(err)
	}
	var name, namespace string = "hello-world-test", "agilis-dev"
	defer client.CancelTask(name)
	if _, err := client.StartTask(name, "task-123", "python"); err != nil {
		t.Fatal(err)
	}
}
