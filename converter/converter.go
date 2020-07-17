package converter

import (
	"encoding/json"
	"fmt"

	"strings"

	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/cloud/audit"

	"github.com/golang/protobuf/jsonpb"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"

	log "github.com/sirupsen/logrus"
)

func ConvertLogEntrytoAuditEvent(logEntry *logging.Entry, auditPayload *audit.AuditLog) (*auditv1.Event, error) {

	m := &jsonpb.Marshaler{}

	log.Debugf("In ConvertLogEntrytoAuditEvent()")
	log.Tracef("Will try to convert: logEntry=%+v, auditPayload=%+v\n", logEntry, auditPayload)

	methodNameParts := strings.Split(auditPayload.MethodName, ".")

	resourceNameParts := strings.Split(auditPayload.ResourceName, "/")
	timestampMicro := metav1.NewMicroTime(logEntry.Timestamp)

	// By default assume 201 status when the verb is created, 200 otherwise.
	// This may be replaced with a 4xx status based on the type of the response
	// and/or the subresource.
	verb := methodNameParts[len(methodNameParts)-1]
	var status *metav1.Status

	if verb == "create" {
		status = &metav1.Status{
			Status:  "Created (inferred)",
			Code:    201,
			Message: "Created (inferred)",
		}

	} else {
		status = &metav1.Status{
			Status:  "OK (inferred)",
			Code:    200,
			Message: "OK (inferred)",
		}
	}

	var ObjectReference *auditv1.ObjectReference

	if len(resourceNameParts) == 6 &&
		resourceNameParts[2] == "namespaces" {
		// The object reference includes a namespace and object name
		ObjectReference = &auditv1.ObjectReference{
			APIGroup:   resourceNameParts[0],
			APIVersion: resourceNameParts[1],
			Namespace:  resourceNameParts[3],
			Resource:   resourceNameParts[4],
			Name:       resourceNameParts[5],
		}
	} else if len(resourceNameParts) == 5 &&
		resourceNameParts[2] == "namespaces" {
		// The object reference does include a namespace but does not have an
		// object name
		ObjectReference = &auditv1.ObjectReference{
			APIGroup:   resourceNameParts[0],
			APIVersion: resourceNameParts[1],
			Namespace:  resourceNameParts[3],
			Resource:   resourceNameParts[4],
		}
	} else if len(resourceNameParts) == 4 {
		// The object reference does not include a namespace
		ObjectReference = &auditv1.ObjectReference{
			APIGroup:   resourceNameParts[0],
			APIVersion: resourceNameParts[1],
			Resource:   resourceNameParts[2],
			Name:       resourceNameParts[3],
		}
	} else if len(resourceNameParts) >= 7 &&
		resourceNameParts[2] == "namespaces" &&
		(resourceNameParts[6] == "attach" ||
			resourceNameParts[6] == "exec") {
		// The object reference includes a subresource. For now, only doing this
		// for attach resources (kubectl exec/attach)
		ObjectReference = &auditv1.ObjectReference{
			APIGroup:    resourceNameParts[0],
			APIVersion:  resourceNameParts[1],
			Namespace:   resourceNameParts[3],
			Resource:    resourceNameParts[4],
			Name:        resourceNameParts[5],
			Subresource: resourceNameParts[6],
		}
	} else {
		return nil, fmt.Errorf("Could not create ObjectReference from resource name %s", auditPayload.ResourceName)
	}

	// The level is RequestResponse and stage is ResponseComplete by default. We
	// change it for pod attach/exec actions.

	var level auditv1.Level
	var stage auditv1.Stage

	if ObjectReference.Subresource == "attach" ||
		ObjectReference.Subresource == "exec" {
		level = "Request"
		stage = "ResponseStarted"
		status.Code = 101
		status.Status = "Switching Protocols (inferred)"
		status.Message = "Switching Protocols (inferred)"
	} else {
		level = "RequestResponse"
		stage = "ResponseComplete"
	}

	auditEvent := &auditv1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "audit.k8s.io/v1beta1",
		},
		Level:      level,
		AuditID:    types.UID(logEntry.InsertID),
		ObjectRef:  ObjectReference,
		Stage:      stage,
		RequestURI: auditPayload.ResourceName,
		Verb:       verb,
		User: authv1.UserInfo{
			Username: auditPayload.AuthenticationInfo.PrincipalEmail,
		},
		SourceIPs:                []string{auditPayload.RequestMetadata.CallerIp},
		ResponseStatus:           status,
		RequestReceivedTimestamp: timestampMicro,
		StageTimestamp:           timestampMicro,
		Annotations:              logEntry.Labels,
	}

	if auditPayload.GetRequest() != nil {
		var request runtime.Unknown

		requestJSON, err := m.MarshalToString(auditPayload.GetRequest())
		if err != nil {
			return nil, fmt.Errorf("Could not convert protobuf request to json")
		}

		err = request.UnmarshalJSON([]byte(requestJSON))
		if err != nil {
			return nil, fmt.Errorf("Could not serialize protobuf request json")
		}

		auditEvent.RequestObject = &request
	}

	if auditPayload.GetResponse() != nil {
		var response runtime.Unknown

		responseJSON, err := m.MarshalToString(auditPayload.GetResponse())
		if err != nil {
			return nil, fmt.Errorf("Could not convert protobuf response to json")
		}

		var responseObject map[string]interface{}

		err = json.Unmarshal([]byte(responseJSON), &responseObject)

		if err != nil {
			return nil, fmt.Errorf("Could not unmarshal response json")
		}

		err = response.UnmarshalJSON([]byte(responseJSON))
		if err != nil {
			return nil, fmt.Errorf("Could not serialize protobuf response json")
		}

		auditEvent.ResponseObject = &response

		// If the type of the response is core.k8s.io/v1.Status *and* the status
		// was not "Success", save that to the audit log status. Otherwise,
		// assume the operation succeeded and return a 200 status.
		if responseObject["@type"] == "core.k8s.io/v1.Status" {

			var status metav1.Status

			err = json.Unmarshal([]byte(responseJSON), &status)

			if err != nil {
				return nil, fmt.Errorf("Could not deserialize response as status")
			}

			if status.Status != "Success" {
				auditEvent.ResponseStatus = &status
			}
		}
	}

	return auditEvent, nil
}
