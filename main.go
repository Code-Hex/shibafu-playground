package main

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/datastore"
)

var (
	// version will be set when compiled by Makefile
	version string

	log = newStdLogger()
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error creating server: %v", err)
	}
}

func run() error {
	env, err := NewEnvConfig()
	if err != nil {
		return err
	}

	pid := projectID()
	db, err := getDBByProjectID(pid)
	if err != nil {
		return err
	}

	srv := NewServer(env.HostName, db, log)
	log.Printf("Listening on :%d ...", env.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", env.Port), srv)
}

func projectID() string {
	id, err := metadata.ProjectID()
	log.Errorf("Could not determine the project ID: %v", err)
	return id
}

func getDBByProjectID(pid string) (store, error) {
	if pid == "" {
		return &inMemStore{}, nil
	}
	c, err := datastore.NewClient(context.Background(), pid)
	if err != nil {
		return nil, fmt.Errorf("could not create cloud datastore client: %v", err)
	}
	return cloudDatastore{client: c}, nil

}
