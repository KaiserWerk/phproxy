package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type (
	AppConfig struct {
		PHProxy      PHProxy  `yaml:"phproxy"`
		Apps         []PHPApp `yaml:"apps"`
		originalFile string
		mut          *sync.Mutex
	}
	PHProxy struct {
		BindAddr string `yaml:"bind_addr"`
		TLS      TLS    `yaml:"tls"`
	}
	TLS struct {
		KeyFile     string      `yaml:"key_file"`
		CertFile    string      `yaml:"cert_file"`
		LetsEncrypt LetsEncrypt `yaml:"letsencrypt"`
	}
	LetsEncrypt struct {
		Enabled   bool   `yaml:"enabled"`
		Domain    string `yaml:"domain"` // []string?
		Email     string `yaml:"email"`
		AcceptTOS bool   `yaml:"accept_tos"`
	}

	PHPApp struct {
		Name         string       `yaml:"name"`
		Host         string       `yaml:"host"`
		Caching      Caching      `yaml:"caching"`
		UrlRewriting UrlRewriting `yaml:"url_rewriting"`
	}
	Caching struct {
		Level        uint8    `yaml:"level"`
		StaticAssets []string `yaml:"static_assets"`
	}
	UrlRewriting struct {
		Custom        map[string]string `yaml:"custom"`
		HtaccessFiles []string          `yaml:"htaccess_files"`
	}
)

func Init(file string) (*AppConfig, error) {
	cfg := AppConfig{}

	cont, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(cont, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *AppConfig) ToFile(file string) error {
	// mutex
	// orig name

	y, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(file, y, 0644)
}
