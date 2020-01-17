package converter_test

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/converter"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/model"
	"google.golang.org/genproto/googleapis/cloud/audit"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"

	"github.com/stretchr/testify/assert"
)

func TestConversions(t *testing.T) {

	testFilesDir := "./test_files"
	logEntriesDir := path.Join(testFilesDir, "log_entries")
	auditEvtsDir := path.Join(testFilesDir, "k8s_audit_events")

	files, err := ioutil.ReadDir(logEntriesDir)
	if err != nil {
		t.Fatalf("Could not read directory containing log events: %v", err)
	}

	for _, file := range files {

		t.Logf("Test File: %s", file.Name())
		content, err := ioutil.ReadFile(path.Join(logEntriesDir, file.Name()))
		if err != nil {
			t.Fatalf("Could not read contents of log entries file %s: %v", file.Name(), err)
		}

		content = []byte(strings.TrimSpace(string(content)))

		var entry model.SavedLoggingEntry
		err = json.Unmarshal(content, &entry)

		if err != nil {
			t.Fatalf("Could not decode log entry as json: %v", err)
		}

		var auditPayload audit.AuditLog
		err = jsonpb.UnmarshalString(entry.AuditPayload, &auditPayload)

		if err != nil {
			t.Fatalf("Could not decode audit payload: %v", err)
		}

		t.Logf("Log Event: %+v", &entry)
		t.Logf("Audit Payload: %+v", auditPayload)

		actualAuditEvent, err := converter.ConvertLogEntrytoAuditEvent(entry.Entry, &auditPayload)
		if err != nil {
			t.Fatalf("Could not convert log entry to audit object: %v", err)
		}

		var expectedAuditPayload auditv1.Event

		expectedAuditContent, err := ioutil.ReadFile(path.Join(auditEvtsDir, file.Name()))
		if err != nil {
			t.Fatalf("Could not read contents of audit events file %s: %v", file.Name(), err)
		}

		t.Logf("expected STR: %s", expectedAuditContent)

		err = json.Unmarshal(expectedAuditContent, &expectedAuditPayload)
		if err != nil {
			t.Fatalf("Could not convert expected audit event to json struct: %v", err)
		}

		expectedAuditEventJSON, err := json.MarshalIndent(expectedAuditPayload, "", "  ")
		if err != nil {
			t.Fatalf("Could not convert expected audit event to json struct: %v", err)
		}

		actualAuditEventJSON, err := json.MarshalIndent(*actualAuditEvent, "", "  ")
		if err != nil {
			t.Fatalf("Could not marshal actual audit event json: %v", err)
		}

		assert.EqualValues(t, string(expectedAuditEventJSON), string(actualAuditEventJSON), "*******%s: Expected and Actual K8s Audit Event Differ", file.Name())
	}

}
