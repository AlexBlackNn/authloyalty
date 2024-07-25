package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

type RedisSentinelConfig struct {
	MasterName     string `yaml:"masterName"`
	SentinelAddrs1 string `yaml:"sentinelAddrs1"`
	SentinelAddrs2 string `yaml:"sentinelAddrs2"`
	SentinelAddrs3 string `yaml:"sentinelAddrs3"`
	Password       string `yaml:"password"`
}

type StoragePatroniConfig struct {
	Master string `yaml:"master"`
	Slave  string `yaml:"slave"`
}

type Config struct {
	// without this param will be used "local" as param value
	Env             string        `yaml:"env" env-default:"local"`
	AccessTokenTtl  time.Duration `yaml:"access_token_ttl"  env-required:"true"`
	RefreshTokenTtl time.Duration `yaml:"refresh_token_ttl"  env-required:"true"`
	RedisAddress    string        `yaml:"redis_address"`
	// without this param can't work
	StoragePath    string               `yaml:"storage_path"`
	ServiceSecret  string               `yaml:"service_secret" env-required:"true"`
	GRPC           GRPCConfig           `yaml:"grpc"`
	RedisSentinel  RedisSentinelConfig  `yaml:"redis_sentinel"`
	StoragePatroni StoragePatroniConfig `yaml:"storage_patroni"`
	JaegerUrl      string               `yaml:"jaeger_url"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exists: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config " + err.Error())
	}
	return &cfg
}

// fetchConfigPath fetches config path from command line flag or env var
// Priority: flag -> env -> default
// Default value is empty string
func fetchConfigPath() string {
	var res string
	// --config="path/to/config.yaml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
