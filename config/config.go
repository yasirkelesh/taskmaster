package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v3"
)

type Program struct {
	Name        string   `yaml:"name"`
	Cmd         string   `yaml:"cmd"`
	NumProcs    int      `yaml:"numprocs"`
	AutoStart   bool     `yaml:"autostart"`
	AutoRestart string   `yaml:"autorestart"`
	ExitCodes   []int    `yaml:"exitcodes"`
	StartTime   int      `yaml:"starttime"`
	StopSignal  string   `yaml:"stopsignal"`
	StopTime    int      `yaml:"stoptime"`
	Env         []string `yaml:"env"`
	WorkingDir  string   `yaml:"workingdir"`
}

type Config struct {
	Programs []Program `yaml:"programs"`
}

func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
