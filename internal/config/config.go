package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Storage     StorageConfig     `mapstructure:"storage"`
	Kubernetes  KubeConfig        `mapstructure:"kubernetes"`
	Detector    DetectorConfig    `mapstructure:"detector"`
	LLM         LLMConfig         `mapstructure:"llm"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Chaos       ChaosConfig       `mapstructure:"chaos"`
	Remediation RemediationConfig `mapstructure:"remediation"`
}

type ServerConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	LogLevel    string        `mapstructure:"log_level"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

type KubeConfig struct {
	Mode           string `mapstructure:"mode"`
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
	FakeNamespace  string `mapstructure:"fake_namespace"`
}

type DetectorConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	CheckInterval  time.Duration `mapstructure:"check_interval"`
	PodWatch       bool          `mapstructure:"pod_watch"`
	NodeWatch      bool          `mapstructure:"node_watch"`
	Namespaces     []string      `mapstructure:"namespaces"`
	MinSeverity    string        `mapstructure:"min_severity"`
	CooldownPeriod time.Duration `mapstructure:"cooldown_period"`
}

type LLMConfig struct {
	Provider    string        `mapstructure:"provider"`
	APIKey      string        `mapstructure:"api_key"`
	Model       string        `mapstructure:"model"`
	BaseURL     string        `mapstructure:"base_url"`
	MaxTokens   int           `mapstructure:"max_tokens"`
	Temperature float64       `mapstructure:"temperature"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type ChaosConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	AllowedNamespaces []string `mapstructure:"allowed_namespaces"`
	RequireDangerous  bool     `mapstructure:"require_dangerous"`
}

type RemediationConfig struct {
	AutoApprove    bool     `mapstructure:"auto_approve"`
	MaxRetries     int      `mapstructure:"max_retries"`
	AllowedActions []string `mapstructure:"allowed_actions"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:        "0.0.0.0",
			Port:        8080,
			ReadTimeout: 30 * time.Second,
			LogLevel:    "info",
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: "./data/polaris.db",
		},
		Kubernetes: KubeConfig{
			Mode:          "auto",
			KubeconfigPath: "",
			FakeNamespace: "default",
		},
		Detector: DetectorConfig{
			Enabled:        true,
			CheckInterval:  30 * time.Second,
			PodWatch:       true,
			NodeWatch:      true,
			Namespaces:     []string{""},
			MinSeverity:    "warning",
			CooldownPeriod: 5 * time.Minute,
		},
		LLM: LLMConfig{
			Provider:    "deepseek",
			Model:       "deepseek-chat",
			MaxTokens:   2048,
			Temperature: 0.1,
			Timeout:     60 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Chaos: ChaosConfig{
			Enabled:           true,
			AllowedNamespaces: []string{"default"},
			RequireDangerous:  true,
		},
		Remediation: RemediationConfig{
			AutoApprove:    false,
			MaxRetries:     3,
			AllowedActions: []string{"restart", "scale_up", "rollback"},
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("polaris")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("POLARIS")
	viper.AutomaticEnv()

	_ = viper.BindEnv("server.port", "POLARIS_SERVER_PORT")
	_ = viper.BindEnv("kubernetes.mode", "POLARIS_KUBERNETES_MODE")
	_ = viper.BindEnv("llm.api_key", "POLARIS_LLM_API_KEY")
	_ = viper.BindEnv("llm.model", "POLARIS_LLM_MODEL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
