package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type NodeConfig struct {
	Name             string `yaml:"name"`
	DownAfterNoAlive int    `yaml:"down_after_noalive"`
	IdleConns        int    `yaml:"idle_conns"`
	RWSplit          bool   `yaml:"rw_split"`

	User     string `yaml:"user"`
	Password string `yaml:"password"`

	Master string `yaml:"master"`
	Slave  string `yaml:"slave"`
}

type SchemaConfig struct {
	DB    string `yaml:"db"`
	Node  string `yaml:"node"`
	Auths []Auth `yaml:"auths"`
}

type Auth struct {
	User   string `yaml:"user"`
	Passwd string `yaml:"passwd"`
}

type Config struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	LogLevel string `yaml:"log_level"`
	ReqRate  int64  `yaml:"rate"`
	ReqBurst int64  `yaml:"burst"`

	Nodes []NodeConfig `yaml:"nodes"`

	Schemas []SchemaConfig `yaml:"schemas"`
}

func ParseConfigData(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}

	if cfg.ReqRate > 0 {
		/* integer value, 1 corresponds to 0.001 r/s */
		cfg.ReqRate = cfg.ReqRate * 1000
	}

	if cfg.ReqBurst > 0 {
		/* integer value, 1 corresponds to 0.001 r/s */
		cfg.ReqBurst = cfg.ReqBurst * 1000
	}
	return &cfg, nil
}

func ParseConfigFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return ParseConfigData(data)
}
