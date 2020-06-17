package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func start(c string) (func(), error) {
	args := strings.Fields(c)
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	log.Printf("Framework server started.")

	// Give it a second to do its setup.
	time.Sleep(time.Second)

	shutdown := func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill process: ", err)
		}
		log.Printf("Framework server shut down.")
	}
	return shutdown, nil
}

func sendHTTP(url, data string) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading HTTP response body: %v", err)
		}
		return fmt.Errorf("validation failed with exit code %d: %v", resp.StatusCode, string(body))
	}
	return nil
}

func sendCE(url string, e cloudevents.Event) error {
	ctx := cloudevents.ContextWithTarget(context.Background(), url)

	p, err := cloudevents.NewHTTP()
	if err != nil {
		return fmt.Errorf("failed to create protocol: %v", err)
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		return fmt.Errorf("failed to create client, %v", err)
	}

	res := c.Send(ctx, e)
	if !cloudevents.IsACK(res) {
		return fmt.Errorf("failed to send CloudEvent: %v", res)
	}
	return nil
}