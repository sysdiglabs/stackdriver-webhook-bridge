package config_test

import (
	"testing"
	"time"

	pflag "github.com/spf13/pflag"

	"github.com/stretchr/testify/assert"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg, err := config.New("", nil)

	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit", cfg.Url)
	assert.Equal(t, "", cfg.ProjectId)
	assert.Equal(t, "", cfg.ClusterName)
	assert.Equal(t, "", cfg.OutfileName)
	assert.Equal(t, "", cfg.LogfileName)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 30*time.Second, cfg.LagInterval)
	assert.Equal(t, 100, cfg.MaxAuditEventsBatch)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, false, cfg.SupressObjectConversionErrors)
}

func TestConfigCommandLineArgsAllArgs(t *testing.T) {

	testArgs := pflag.NewFlagSet("test", pflag.ExitOnError)

	testArgs.String("url", "", "")
	testArgs.String("project", "", "")
	testArgs.String("cluster", "", "")
	testArgs.String("logfile", "", "")
	testArgs.String("outfile", "", "")
	testArgs.Duration("poll_interval", 1*time.Second, "")
	testArgs.Duration("lag_interval", 1*time.Second, "")
	testArgs.String("log_level", "", "")
	testArgs.String("supress_object_conversion_errors", "", "")

	assert.Nil(t, testArgs.Parse([]string{
		"--url", "some-fake-url",
		"--project", "some-fake-project",
		"--cluster", "some-fake-cluster",
		"--logfile", "some-fake-logfile",
		"--outfile", "some-fake-outfile",
		"--poll_interval", "10s",
		"--lag_interval", "20s",
		"--log_level", "debug",
		"--supress_object_conversion_errors", "true",
	}))

	cfg, err := config.New("", testArgs)

	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "some-fake-url", cfg.Url)
	assert.Equal(t, "some-fake-project", cfg.ProjectId)
	assert.Equal(t, "some-fake-cluster", cfg.ClusterName)
	assert.Equal(t, "some-fake-outfile", cfg.OutfileName)
	assert.Equal(t, "some-fake-logfile", cfg.LogfileName)
	assert.Equal(t, 10*time.Second, cfg.PollInterval)
	assert.Equal(t, 20*time.Second, cfg.LagInterval)
	assert.Equal(t, 100, cfg.MaxAuditEventsBatch)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, true, cfg.SupressObjectConversionErrors)
}

func TestConfigCommandLineArgsNoArgs(t *testing.T) {

	testArgs := pflag.NewFlagSet("test", pflag.ExitOnError)

	testArgs.String("url", "", "")
	testArgs.String("project", "", "")
	testArgs.String("cluster", "", "")
	testArgs.String("logfile", "", "")
	testArgs.String("outfile", "", "")
	testArgs.Duration("poll_interval", 1*time.Second, "")
	testArgs.Duration("lag_interval", 1*time.Second, "")
	testArgs.String("log_level", "", "")

	// Noting that with no args actually set, you retain the defaults
	assert.Nil(t, testArgs.Parse([]string{}))

	cfg, err := config.New("", testArgs)

	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit", cfg.Url)
	assert.Equal(t, "", cfg.ProjectId)
	assert.Equal(t, "", cfg.ClusterName)
	assert.Equal(t, "", cfg.OutfileName)
	assert.Equal(t, "", cfg.LogfileName)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 30*time.Second, cfg.LagInterval)
	assert.Equal(t, 100, cfg.MaxAuditEventsBatch)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, false, cfg.SupressObjectConversionErrors)
}

func TestConfigFilePrecedence(t *testing.T) {

	cfg, err := config.New("./test", nil)

	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "my-file-url", cfg.Url)
	assert.Equal(t, "my-file-project", cfg.ProjectId)
	assert.Equal(t, "my-file-cluster", cfg.ClusterName)
	assert.Equal(t, "my-file-outfile", cfg.OutfileName)
	assert.Equal(t, "my-file-logfile", cfg.LogfileName)
	assert.Equal(t, 21*time.Second, cfg.PollInterval)
	assert.Equal(t, 48*time.Second, cfg.LagInterval)
	assert.Equal(t, 100, cfg.MaxAuditEventsBatch)
	assert.Equal(t, "warning", cfg.LogLevel)
}

func TestConfigFileNoFile(t *testing.T) {

	cfg, err := config.New("./test-noexist", nil)

	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit", cfg.Url)
	assert.Equal(t, "", cfg.ProjectId)
	assert.Equal(t, "", cfg.ClusterName)
	assert.Equal(t, "", cfg.OutfileName)
	assert.Equal(t, "", cfg.LogfileName)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 30*time.Second, cfg.LagInterval)
	assert.Equal(t, 100, cfg.MaxAuditEventsBatch)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfigCmdlinePrecedence(t *testing.T) {

	testArgs := pflag.NewFlagSet("test", pflag.ExitOnError)

	testArgs.String("url", "", "")
	testArgs.String("project", "", "")
	testArgs.String("cluster", "", "")
	testArgs.String("logfile", "", "")
	testArgs.String("outfile", "", "")
	testArgs.Duration("poll_interval", 1*time.Second, "")
	testArgs.Duration("lag_interval", 1*time.Second, "")
	testArgs.String("log_level", "", "")

	assert.Nil(t, testArgs.Parse([]string{
		"--url", "some-fake-url",
		"--project", "some-fake-project",
		"--cluster", "some-fake-cluster",
		"--logfile", "some-fake-logfile",
		"--outfile", "some-fake-outfile",
		"--log_level", "debug",
	}))

	cfg, err := config.New("./test", nil)

	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "my-file-url", cfg.Url)
	assert.Equal(t, "my-file-project", cfg.ProjectId)
	assert.Equal(t, "my-file-cluster", cfg.ClusterName)
	assert.Equal(t, "my-file-outfile", cfg.OutfileName)
	assert.Equal(t, "my-file-logfile", cfg.LogfileName)
	assert.Equal(t, 21*time.Second, cfg.PollInterval)
	assert.Equal(t, 48*time.Second, cfg.LagInterval)
	assert.Equal(t, 100, cfg.MaxAuditEventsBatch)
	assert.Equal(t, "warning", cfg.LogLevel)
}
