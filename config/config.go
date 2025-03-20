package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config, yapılandırma dosyasındaki tüm ayarları tutar

type Config struct {
	Programs map[string]Program `yaml:"programs"`
}

type Program struct {
	Command      string            `yaml:"command"`
	NumProcs     int               `yaml:"numprocs"`
	AutoStart    bool              `yaml:"autostart"`
	AutoRestart  string            `yaml:"autorestart"`
	ExitCodes    []int             `yaml:"exitcodes"`
	StartSecs    int               `yaml:"startsecs"`
	StartRetries int               `yaml:"startretries"`
	StopSignal   string            `yaml:"stopsignal"`
	StopTime     int               `yaml:"stoptime"`
	Stdout       string            `yaml:"stdout"`
	Stderr       string            `yaml:"stderr"`
	Env          map[string]string `yaml:"env"`
	WorkingDir   string            `yaml:"workingdir"`
	Umask        int               `yaml:"umask"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}