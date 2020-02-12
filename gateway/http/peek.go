package http

// func getLatestFromProjectAndStream(w http.ResponseWriter, r *http.Request) error {
// 	projectName := toBackendName(chi.URLParam(r, "projectName"))
// 	streamName := toBackendName(chi.URLParam(r, "streamName"))
// 	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
// 	if instanceID == uuid.Nil {
// 		return httputil.NewError(404, "instance for stream not found")
// 	}

// 	return getLatestFromInstanceID(w, r, instanceID)
// }

// func getLatestFromInstance(w http.ResponseWriter, r *http.Request) error {
// 	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
// 	if err != nil {
// 		return httputil.NewError(404, "instance not found -- malformed ID")
// 	}

// 	return getLatestFromInstanceID(w, r, instanceID)
// }

// func getLatestFromInstanceID(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
// 	// get auth
// 	secret := middleware.GetSecret(r.Context())

// 	// get cached stream
// 	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
// 	if stream == nil {
// 		return httputil.NewError(404, "stream not found")
// 	}

// 	// set log payload
// 	payload := peekTags{
// 		InstanceID: instanceID.String(),
// 	}
// 	middleware.SetTagsPayload(r.Context(), payload)

// 	// check allowed to read stream
// 	perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.Public, stream.External)
// 	if !perms.Read {
// 		return httputil.NewError(403, "secret doesn't grant right to read this stream")
// 	}

// 	// check isn't batch
// 	if stream.Batch {
// 		return httputil.NewError(400, "cannot get latest records for batch streams")
// 	}

// 	// check quota
// 	if !secret.IsAnonymous() {
// 		usage := gateway.Metrics.GetCurrentUsage(r.Context(), secret.GetOwnerID())
// 		ok := secret.CheckReadQuota(usage)
// 		if !ok {
// 			return httputil.NewError(429, "you have exhausted your monthly quota")
// 		}
// 	}

// 	// read body
// 	var body map[string]interface{}
// 	err := jsonutil.Unmarshal(r.Body, &body)
// 	if err == io.EOF {
// 		// no body -- try reading from url parameters
// 		body = make(map[string]interface{})

// 		// read limit
// 		if limit := r.URL.Query().Get("limit"); limit != "" {
// 			body["limit"] = limit
// 		}

// 		// read before
// 		if before := r.URL.Query().Get("before"); before != "" {
// 			body["before"] = before
// 		}
// 	} else if err != nil {
// 		return httputil.NewError(400, "couldn't parse body -- is it valid JSON?")
// 	}

// 	// make sure there's no accidental keys
// 	for k := range body {
// 		if k != "limit" && k != "before" {
// 			return httputil.NewError(400, "unrecognized query key '%s'; valid keys are 'limit' and 'before'", k)
// 		}
// 	}

// 	// get limit
// 	limit, err := parseLimit(body["limit"])
// 	if err != nil {
// 		return httputil.NewError(400, err.Error())
// 	}
// 	payload.Limit = int32(limit)

// 	// get before
// 	before, err := timeutil.Parse(body["before"], true)
// 	if err != nil {
// 		return httputil.NewError(400, err.Error())
// 	}
// 	if !before.IsZero() {
// 		payload.Before = timeutil.UnixMilli(before)
// 	}

// 	// update payload after setting new details
// 	middleware.SetTagsPayload(r.Context(), payload)

// 	// prepare write (we'll be writing as we get data, not in one batch)
// 	w.Header().Set("Content-Type", "application/json")

// 	// begin json object
// 	result := make([]interface{}, 0, limit)
// 	bytesRead := 0

// 	// read rows from engine
// 	err = hub.Engine.Tables.ReadLatestRecords(r.Context(), instanceID, limit, before, func(avroData []byte, timestamp time.Time) error {
// 		// decode avro
// 		data, err := stream.Codec.UnmarshalAvro(avroData)
// 		if err != nil {
// 			return err
// 		}

// 		// convert to json friendly
// 		data, err = stream.Codec.ConvertFromAvroNative(data, true)
// 		if err != nil {
// 			return err
// 		}

// 		// set timestamp
// 		data["@meta"] = map[string]int64{"timestamp": timeutil.UnixMilli(timestamp)}

// 		// done
// 		result = append(result, data)
// 		bytesRead += len(avroData)
// 		return nil
// 	})
// 	if err != nil {
// 		return httputil.NewError(400, err.Error())
// 	}

// 	// prepare result for encoding
// 	encode := map[string]interface{}{"data": result}

// 	// write and finish
// 	w.Header().Set("Content-Type", "application/json")
// 	err = jsonutil.MarshalWriter(encode, w)
// 	if err != nil {
// 		return err
// 	}

// 	// track read metrics
// 	gateway.Metrics.TrackRead(stream.StreamID, int64(len(result)), int64(bytesRead))
// 	if !secret.IsAnonymous() {
// 		gateway.Metrics.TrackRead(secret.GetOwnerID(), int64(len(result)), int64(bytesRead))
// 	}

// 	return nil
// }
