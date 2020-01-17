package model

import (
	"time"

	"cloud.google.com/go/logging"
)

// We define this struct instead of simply serializing logging.Entry structs
// because:
// 1. logging.Entry.payload is a generic interface{} and when cast it
//    loses any notion of the type it had.
// 2. For the logs we care about (k8s audit logs), the
//    payload is an audit.AuditLog struct and that can't be serialized using
//    encoder/json.
type SavedLoggingEntry struct {
	Entry *logging.Entry

	// This should be an audit.AuditLog struct, serialized to a string using
	// https://godoc.org/github.com/golang/protobuf/jsonpb's Marshaler
	AuditPayload string
}

type Config struct {
	Url                 string
	ProjectArg          string
	OutfileName         string
	LogfileName         string
	PollInterval        time.Duration
	MaxAuditEventsBatch int
}

func NewConfig() *Config {
	return &Config{
		Url:                 "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit",
		ProjectArg:          "",
		OutfileName:         "",
		LogfileName:         "",
		PollInterval:        5,
		MaxAuditEventsBatch: 100,
	}
}
