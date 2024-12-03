package config

import "errors"

var (
	ErrEmptyConfigPath  = errors.New("config path is empty")
	ErrAbsentConfigFile = errors.New("config file does not exists")
	ErrReadConfigFailed = errors.New("reading config file failed")
)
