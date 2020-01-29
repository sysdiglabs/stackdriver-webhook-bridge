package poller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/logging/logadmin"

	"github.com/golang/protobuf/jsonpb"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/converter"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/model"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/cloud/audit"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"

	log "github.com/sirupsen/logrus"
)

type Poller struct {
	ctx        context.Context
	client     *logadmin.Client
	httpClient *http.Client
	cfg        *model.Config
	project    string
	marshaler  *jsonpb.Marshaler
	logfile    *os.File
	outfile    *os.File
}

func NewPoller(ctx context.Context, cfg *model.Config) (*Poller, error) {

	p := &Poller{
		ctx:        ctx,
		cfg:        cfg,
		httpClient: &http.Client{},
		marshaler:  &jsonpb.Marshaler{},
	}

	var err error
	if cfg.ProjectArg != "" {
		log.Infof("Using project id from config: %s", cfg.ProjectArg)
		p.project = cfg.ProjectArg
	} else {
		log.Debugf("Project blank, using metadata service to find project name...")

		url := "http://metadata.google.internal/computeMetadata/v1/project/project-id"
		req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))

		if err != nil {
			return nil, fmt.Errorf("Could not construct http request to %s: %v", url, err)
		}

		req.Header.Set("Metadata-Flavor", "Google")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Could not GET metadata from %s: %v", url, err)
		}

		if err != nil {
			return nil, fmt.Errorf("Error fetching project id from metadata service: %v", err)
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Non-200 response fetching project id from metadata service: status=%s body=%s", resp.Status, body)
		}

		log.Infof("Using project id from metadata service: %s", body)
		p.project = string(body)
	}

	if cfg.LogfileName != "" {
		log.Infof("Will append log entries to: %s", cfg.LogfileName)
		p.logfile, err = os.OpenFile(cfg.LogfileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("Could not open %s for writing: %v", cfg.LogfileName, err)
		}
	}

	if cfg.OutfileName != "" {
		log.Infof("Will append audit events to: %s", cfg.OutfileName)
		p.outfile, err = os.OpenFile(cfg.OutfileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("Could not open %s for writing: %v", cfg.OutfileName, err)
		}
	}

	p.client, err = logadmin.NewClient(ctx, p.project)
	if err != nil {
		return nil, fmt.Errorf("Could not create log reader: %v", err)
	}

	log.Infof("Will read events from project id: %s", p.project)
	log.Infof("Will post events to webhook: %s", cfg.Url)

	return p, nil
}

func (p *Poller) Close() {

	log.Infof("Closing log reader...")

	if err := p.client.Close(); err != nil {
		log.Errorf("Could not close log reader: %v", err)
	}
}

func (p *Poller) sendAuditEventsBatch(auditEvents []*auditv1.Event) error {

	auditEventsJSON, err := json.Marshal(auditEvents)
	if err != nil {
		return fmt.Errorf("Could not serialize audit events to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", p.cfg.Url, bytes.NewBuffer(auditEventsJSON))

	if err != nil {
		return fmt.Errorf("Could not construct http request to %s: %v", p.cfg.Url, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Could not POST audit events to %s: %v", p.cfg.Url, err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Debugf("response from post: status=%s body=%s:", resp.Status, string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response %s from POST of audit events: %s", resp.Status, string(body))
	}
	log.Infof("Forwarded %d events", len(auditEvents))

	return nil
}

func (p *Poller) PollLogsSendEvents(curTime time.Time) time.Time {

	timeStr := curTime.Format(time.RFC3339)
	filter := fmt.Sprintf("logName=\"projects/%s/logs/cloudaudit.googleapis.com%%2Factivity\" AND resource.type=\"k8s_cluster\" AND timestamp >= \"%s\"", p.project, timeStr)
	it := p.client.Entries(p.ctx, logadmin.Filter(filter))

	log.Debugf("Fetching all logs since %v, filter=%s...", curTime, filter)

	var entryStr []byte

	var auditEvents []*auditv1.Event

	for entry, err := it.Next(); err != iterator.Done; entry, err = it.Next() {

		log.Debugf("Got log entry: %+v", entry)

		if entry == nil {
			// Just prevents runaway loop in case of misconfiguration
			time.Sleep(1 * time.Second)
			continue
		}

		// The filter only has second-level resolution, so check the timestamp exactly
		if !entry.Timestamp.After(curTime) {
			continue
		}

		curTime = entry.Timestamp

		var auditPayload *audit.AuditLog
		var ok bool
		if auditPayload, ok = entry.Payload.(*audit.AuditLog); !ok {
			log.Errorf("Could not extract payload as audit payload")
			continue
		}

		if p.logfile != nil {
			auditStr, err := p.marshaler.MarshalToString(auditPayload)

			if err != nil {
				log.Errorf("Could not serialize audit payload: %v", err)
				continue
			}

			savedLogEntry := &model.SavedLoggingEntry{
				Entry:        entry,
				AuditPayload: auditStr,
			}

			entryStr, err = json.Marshal(savedLogEntry)
			if err != nil {
				log.Errorf("Could not convert log entry to json string: %v", err)
				continue
			}

			if p.logfile != nil {
				log.Debugf("saving log entry string: %s", string(entryStr))

				entryStr = append(entryStr, '\n')

				_, err = p.logfile.Write(entryStr)
				if err != nil {
					log.Errorf("Could not write log entry to file %s: %v", p.cfg.LogfileName, err)
					continue
				}
			}
		}

		auditEvent, err := converter.ConvertLogEntrytoAuditEvent(entry, auditPayload)
		if err != nil {
			log.Errorf("Could not convert log entry to audit object: %v", err)
			continue
		}
		auditStr, err := json.Marshal(auditEvent)
		if err != nil {
			log.Errorf("Could not serialize audit object: %v", err)
			continue
		}
		log.Debugf("Got audit event: %s", string(auditStr))

		if p.outfile != nil {
			auditStr = append(auditStr, '\n')
			_, err = p.outfile.Write(auditStr)
			if err != nil {
				log.Errorf("Could not write audit event to file %s: %v", p.cfg.OutfileName, err)
				continue
			}
		}

		auditEvents = append(auditEvents, auditEvent)

		if len(auditEvents) >= p.cfg.MaxAuditEventsBatch {
			err = p.sendAuditEventsBatch(auditEvents)
			auditEvents = nil
			if err != nil {
				log.Errorf("Could not send batch of audit events: %v", err)
				continue
			}
		}
	}

	if len(auditEvents) > 0 {
		err := p.sendAuditEventsBatch(auditEvents)
		auditEvents = nil
		if err != nil {
			log.Errorf("Could not send batch of audit events: %v", err)
		}
	}

	return curTime
}
