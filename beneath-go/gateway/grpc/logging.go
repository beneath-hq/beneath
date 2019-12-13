package grpc

type writeRecordsLog struct {
	InstanceID   string `json:"instance_id,omitempty"`
	RecordsCount int    `json:"records,omitempty"`
	BytesWritten int    `json:"bytes,omitempty"`
}

type readRecordsLog struct {
	InstanceID string `json:"instance_id,omitempty"`
	Offset     int64  `json:"offset,omitempty"`
	Limit      int32  `json:"limit,omitempty"`
}

type lokupRecordsLog struct {
	InstanceID string `json:"instance_id,omitempty"`
	Limit      int32  `json:"limit,omitempty"`
	Where      string `json:"where,omitempty"`
	After      string `json:"after,omitempty"`
}

type clientPingLog struct {
	ClientID      string `json:"client_id,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
}

type streamDetailsLog struct {
	Stream  string `json:"stream"`
	Project string `json:"project"`
}
