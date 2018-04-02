package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

type Config struct {
	BuildkiteToken  string
	PipelineConfigs []PipelineConfig
}

type PipelineConfig struct {
	Name   string
	Config struct {
		InstanceType string
		KeypairName  string
		StorageSize  int
		AMI          string `json:"ami"`
		UserData     string
		Timeout      Duration
		MaxInstances int
	}
}

func (c Config) GetPipelineConfig(name string) (PipelineConfig, error) {
	var defConf *PipelineConfig
	for i := range c.PipelineConfigs {
		pc := c.PipelineConfigs[i]
		if pc.Name == name {
			return pc, nil
		}
		if pc.Name == "" {
			defConf = &pc
		}
	}
	if defConf == nil {
		return PipelineConfig{}, fmt.Errorf("no pipeline config with name %q", name)
	}
	return *defConf, nil
}

func New() (Config, error) {
	var c Config
	config, err := ioutil.ReadFile("config.json")
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(config, &c)
	if err != nil {
		return c, err
	}
	if err := c.Validate(); err != nil {
		return c, err
	}
	return c, nil
}

func (c Config) Validate() error {
	if c.BuildkiteToken == "" {
		return fmt.Errorf("BuildkiteToken was not set")
	}
	return nil
}
