package poller

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "swb"
	subsystem = "poller"
)

var (
	promLogFetchError                         prometheus.Counter

	promLogEntryIn                            prometheus.Counter
	promAuditEventOut                         prometheus.Counter

	promAuditPayloadExtractError              prometheus.Counter
	promAuditPayloadConvertError              prometheus.Counter
	promAuditEventMarshalError                prometheus.Counter
	promAuditEventSendError                   prometheus.Counter
)

func CreateMetrics() {
	promLogFetchError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "log_fetch_error",
			Help:      "the number of times the bridge had an error fetching a set of stackdriver logs",
		},
	)

	promLogEntryIn = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "log_entry_in",
			Help:      "the number of log entries received",
		},
	)

	promAuditEventOut = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "audit_event_out",
			Help:      "the number of audit events successfully passed along to the agent",
		},
	)

	promAuditPayloadExtractError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "audit_payload_extract_error",
			Help:      "the number of times the bridge had an error extracting the audit payload from a log entry",
		},
	)

	promAuditPayloadConvertError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "audit_payload_convert_error",
			Help:      "the number of times the bridge had an error converting an audit payload to an audit event",
		},
	)

	promAuditEventMarshalError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "audit_event_marshal_error",
			Help:      "the number of times the bridge had an error marshaling an audit event to a json string",
		},
	)

	promAuditEventSendError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "audit_event_send_error",
			Help:      "the number of audit events that could not successfully be sent to the agent",
		},
	)

	prometheus.MustRegister(promLogFetchError)
	prometheus.MustRegister(promLogEntryIn)
	prometheus.MustRegister(promAuditEventOut)
	prometheus.MustRegister(promAuditPayloadExtractError)
	prometheus.MustRegister(promAuditPayloadConvertError)
	prometheus.MustRegister(promAuditEventMarshalError)
	prometheus.MustRegister(promAuditEventSendError)
}

func ResetMetrics() {
	prometheus.Unregister(promLogFetchError)
	prometheus.Unregister(promLogEntryIn)
	prometheus.Unregister(promAuditEventOut)
	prometheus.Unregister(promAuditPayloadExtractError)
	prometheus.Unregister(promAuditPayloadConvertError)
	prometheus.Unregister(promAuditEventMarshalError)
	prometheus.Unregister(promAuditEventSendError)
}

func init() {
	CreateMetrics()
}





