package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

type JWTconfig struct {
	AccessSecret     string `yaml:"access_secret"`
	RefreshSecret    string `yaml:"rafresh_secret"`
	AccessTTLMinutes int    `yaml:"access_ttl_minute"`
	RefreshTTLHours  int    `yaml:"refresh_ttl_hour"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	SMTP   SMTPConfig   `yaml:"smtp"`
	JWT    JWTconfig    `yaml:"jwt"`
}

func LoadConfig(path string) (*Config, error) {

	cfg := &Config{}

	// Read the yaml file
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode the file and assign to the corresponding fields
	if err := yaml.Unmarshal(file, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
