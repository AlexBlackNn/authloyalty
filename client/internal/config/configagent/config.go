package configagent

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AlexBlackNn/authloyalty/client/internal/config"
	"github.com/ilyakaznacheev/cleanenv"
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

// Config consists project settings.
type Config struct {
	Env                   string        `yaml:"env" env-default:"local" env:"ENV"`
	ServerAddr            string        `yaml:"server_addr" env-default:":8080" env:"ADDRESS"`
	PollInterval          int           `yaml:"poll_interval" env-default:"2" env:"POLL_INTERVAL"`
	ReportInterval        int           `yaml:"report_interval" env-default:"10" env:"REPORT_INTERVAL"`
	AgentTimeout          int           `yaml:"client_timeout" env:"CLIENT_TIMEOUT"`
	AgentRetryCount       int           `yaml:"agent_retry_count" env-default:"30" env:"AGENT_RETRY_COUNT" envDefault:"3"`
	AgentRetryWaitTime    time.Duration `yaml:"agent_retry_wait_time" env-default:"30s" env:"AGENT_RETRY_WAIT_TIME" envDefault:"30s"`
	AgentRetryMaxWaitTime time.Duration `yaml:"agent_retry_max_wait_time" env-default:"90s" env:"AGENT_RETRY_MAX_WAIT_TIME" envDefault:"90s"`
	AgentRateLimit        int           `yaml:"agent_rate_limit" env-default:"100" env:"RATE_LIMIT" envDefault:"100"`
	AgentBurstTokens      int           `yaml:"agent_burst_tokens" env-default:"100" env:"AGENT_BURST_TOKENS" envDefault:"100"`
	HashKey               string        `yaml:"hash_key" env:"KEY"`
	CryptoKeyPath         string        `yaml:"env" env-default:"" env:"ENV"`
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"Env: %s,"+
			" ServerAddr: %s, "+
			"PollInterval: %d,"+
			" ReportInterval: %d,"+
			" ClientTimeout: %d",
		c.Env,
		c.ServerAddr,
		c.PollInterval,
		c.ReportInterval,
		c.AgentTimeout,
	)
}

// fetchConfigPath fetches config path from command line flag or env var
// Priority: env -> yml -> flag -> default

// New loads config
func New() (*Config, error) {
	cfg := &Config{}
	var err error
	var configPath string

	flag.StringVar(&cfg.Env, "e", "local", "project environment")
	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "host address")
	flag.StringVar(&cfg.HashKey, "k", "", "hash key")
	flag.IntVar(&cfg.ReportInterval, "r", 2, "metrics report interval")
	flag.IntVar(&cfg.PollInterval, "p", 1, "metrics poll interval")
	flag.IntVar(&cfg.AgentTimeout, "t", 1, "agent request timeout")
	flag.IntVar(&cfg.AgentRateLimit, "l", 100, "agent rate limit")
	flag.StringVar(&configPath, "c", "", "path to config file")
	flag.StringVar(&cfg.CryptoKeyPath, "crypto-key", "", "path to crypto file")

	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	if configPath != "" {
		cfg, err = LoadByPath(configPath)
		if err != nil {
			return &Config{}, err
		}
		return cfg, nil
	}

	err = env.Parse(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// LoadByPath loads config by path
func LoadByPath(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, config.ErrAbsentConfigFile
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return &Config{}, config.ErrReadConfigFailed
	}
	return &cfg, nil
}
