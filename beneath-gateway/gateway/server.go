package gateway

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
)

// GetHandler returns a HTTP handler
func GetHandler() http.Handler {
	handler := chi.NewRouter()

	// TODO: Add graphql
	// GraphQL endpoints
	// handler.Get("/graphql")
	// handler.Get("/projects/{projectName}/graphql")

	// REST endpoints
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}", AppHandler(getFromInstance))
	handler.Method("GET", "/streams/instances/{instanceID}", AppHandler(getFromInstance))
	handler.Method("POST", "/streams/instances/{instanceID}", AppHandler(postToInstance))

	return handler
}

func getFromInstance(w http.ResponseWriter, r *http.Request) error {
	auth, err := parseAuth(r)
	if err != nil {
		return err
	}

	var instanceID uuid.UUID
	instanceIDStr := chi.URLParam(r, "instanceID")
	if instanceIDStr != "" {
		instanceID, err = uuid.FromString(instanceIDStr)
		if err != nil {
			return NewHTTPError(404, "instance not found -- malformed ID")
		}
	} else {
		projectName := chi.URLParam(r, "projectName")
		streamName := chi.URLParam(r, "streamName")
		instanceID, err = lookupCurrentInstanceID(projectName, streamName)
		if err != nil {
			return err
		}
	}

	instance, err := lookupInstance(instanceID)
	if err != nil {
		return err
	}

	role, err := lookupRole(auth, instance)
	if err != nil {
		return err
	}

	if !role.read {
		return NewHTTPError(403, "token doesn't grant right to read this stream")
	}

	// TODO
	// Read from BT in accordance with how we end up writing it
	// Support filter, limit, page (see https://docs.hasura.io/1.0/graphql/manual/queries/query-filters.html)

	w.Write([]byte(fmt.Sprintf("Hello Stream Instance %s", instanceID)))
	return nil
}

func postToInstance(w http.ResponseWriter, r *http.Request) error {
	auth, err := parseAuth(r)
	if err != nil {
		return err
	}

	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return NewHTTPError(404, "instance not found -- malformed ID")
	}

	instance, err := lookupInstance(instanceID)
	if err != nil {
		return err
	}

	role, err := lookupRole(auth, instance)
	if err != nil {
		return err
	}

	if !role.write && !(instance.Manual && role.manage) {
		return NewHTTPError(403, "token doesn't grant right to write to this stream")
	}

	// SPEC
	// - Get schema for stream
	// - Read payload (JSON) and encode with schema
	// - Write to pubsub

	/*
		INPUT
		- instanceID
		- data (json)
		- timestamp: default now
		- seq no: default random

		PAYLOAD TO PUBSUB
		- streamInstanceId
		- ksuid: timestamp + seqno
		- data (binary encoded according to schema)

		INTO BIGQUERY
		- create view beneath.projectName.streamName -> latest streamInstanceId
		- beneath.projectName.streamInstanceId
			- ksuid, timestamp, insert_time, fields in data

		INTO BIGTABLE
		- streamInstanceId#key (key read from data + schema)
		- ksuid, updated_on, fields_in_data (on bigger than previous ksuid)

	*/

	w.Write([]byte(fmt.Sprintf("Wrote Stream Instance %s", instanceID)))
	return nil
}
