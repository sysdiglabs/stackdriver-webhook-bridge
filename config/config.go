package config

import (
	"fmt"
	"os"
	"path"
	"time"

	pflag "github.com/spf13/pflag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Url                 string
	ProjectId           string
	ClusterName         string
	OutfileName         string
	LogfileName         string
	PollInterval        time.Duration
	LagInterval         time.Duration
	MaxAuditEventsBatch int
	PrometheusPort      int
	LogLevel            string
	vcfg                *viper.Viper
}

func New(configDir string, commandLine *pflag.FlagSet) (*Config, error) {

	vcfg := viper.New()

	vcfg.SetDefault("url", "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit")
	vcfg.SetDefault("project", "")
	vcfg.SetDefault("cluster", "")
	vcfg.SetDefault("outfile", "")
	vcfg.SetDefault("logfile", "")
	vcfg.SetDefault("poll_interval", "5s")
	vcfg.SetDefault("lag_interval", "30s")
	vcfg.SetDefault("max-audit-events-batch", 100)
	vcfg.SetDefault("prometheus.port", 25000)
	vcfg.SetDefault("log_level", "info")

	c := &Config{
		vcfg:                      vcfg,
	}

	if configDir != "" {
		err := c.LoadFile(configDir)
		if err != nil {
			return nil, err
		}
	}
	if commandLine != nil {
		err := c.vcfg.BindPFlags(commandLine)
		if err != nil {
			return nil, err
		}
	}

	c.UpdateValues()
	c.LogSettings()

	return c, nil
}

func (c *Config) UpdateValues() {
	c.Url = c.vcfg.GetString("url")
	c.ProjectId = c.vcfg.GetString("project")
	c.ClusterName = c.vcfg.GetString("cluster")
	c.OutfileName = c.vcfg.GetString("outfile")
	c.LogfileName = c.vcfg.GetString("logfile")
	c.PollInterval, _ = time.ParseDuration(c.vcfg.GetString("poll_interval"))
	c.LagInterval, _ = time.ParseDuration(c.vcfg.GetString("lag_interval"))
	c.MaxAuditEventsBatch = c.vcfg.GetInt("max-audit-events-batch")
	c.PrometheusPort = c.vcfg.GetInt("prometheus.port")
	c.LogLevel = c.vcfg.GetString("log_level")
}

func (c *Config) LoadFile(configDir string) error {

	configName := "swb-config"
	configExt := "yaml"

	c.vcfg.SetConfigName(configName)
	c.vcfg.SetConfigType(configExt)

	c.vcfg.AddConfigPath(configDir)

	_, err := os.Stat(path.Join(configDir, fmt.Sprintf("%s.%s", configName, configExt)))

	if err == nil || ! os.IsNotExist(err) {
		if err := c.vcfg.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				return fmt.Errorf("Could not read config file: %v", err)
			} else {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
		}
	}

	c.UpdateValues()

	return nil
}

func (c *Config) LogSettings() {
	log.Info("Current Configuration:")
	for _, key := range c.vcfg.AllKeys() {
		log.Infof(" %s: %v", key, c.vcfg.Get(key))
	}
}
