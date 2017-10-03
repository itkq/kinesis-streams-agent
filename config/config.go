package config

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AggregatorConfig  *AggregatorConfig  `yaml:"aggregator" validate:"required"`
	APIConfig         *APIConfig         `yaml:"api" validate:"required"`
	FileWatcherConfig *FileWatcherConfig `yaml:"watcher" validate:"required"`
	SenderConfig      *SenderConfig      `yaml:"sender" validate:"required"`
	StateConfig       *StateConfig       `yaml:"state" validate:"required"`
}

type AggregatorConfig struct {
	FlushInverval time.Duration `yaml:"flush_interval" validate:"required"`
}

type APIConfig struct {
	Address string `yaml:"address" validate:"required"`
}

type FileWatcherConfig struct {
	LifeTimeAfterMovedFile           time.Duration `yaml:"lifetime_after_file_moved" validate:"required"`
	ReadFileInterval                 time.Duration `yaml:"read_file_interval" validate:"required"`
	UnputtableRecordsLocalBackupPath string        `yaml:"unputtable_record_local_backup_path"`
	WatchPaths                       []string      `yaml:"watch_paths" validate:"required"`
}

type SenderConfig struct {
	ForwardProxyUrl string `yaml:"forward_proxy_url"`
	OutputFilePath  string `yaml:"output_filepath"`
	StreamName      string `yaml:"stream_name"`
	Type            string `yaml:"type" validate:"required"`
}

type StateConfig struct {
	StateFilePath string `yaml:"state_filepath" validate:"required"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	c := Config{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}
