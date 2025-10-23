package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	Payment       PaymentConfig       `mapstructure:"payment"`
	Decorators    DecoratorsConfig    `mapstructure:"decorators"`
	Notifications NotificationsConfig `mapstructure:"notifications"`
	Metrics       MetricsConfig       `mapstructure:"metrics"`
	CLI           CLIConfig           `mapstructure:"cli"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Path            string        `mapstructure:"path"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type PaymentConfig struct {
	Timeout       time.Duration    `mapstructure:"timeout"`
	RetryAttempts int              `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration    `mapstructure:"retry_delay"`
	CreditCard    CreditCardConfig `mapstructure:"credit_card"`
	PayPal        PayPalConfig     `mapstructure:"paypal"`
	Crypto        CryptoConfig     `mapstructure:"crypto"`
}

type CreditCardConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	MinAmount float64 `mapstructure:"min_amount"`
	MaxAmount float64 `mapstructure:"max_amount"`
}

type PayPalConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	MinAmount float64 `mapstructure:"min_amount"`
	MaxAmount float64 `mapstructure:"max_amount"`
}

type CryptoConfig struct {
	Enabled             bool     `mapstructure:"enabled"`
	MinAmount           float64  `mapstructure:"min_amount"`
	MaxAmount           float64  `mapstructure:"max_amount"`
	SupportedCurrencies []string `mapstructure:"supported_currencies"`
}

type DecoratorsConfig struct {
	Discount       DiscountConfig       `mapstructure:"discount"`
	Cashback       CashbackConfig       `mapstructure:"cashback"`
	FraudDetection FraudDetectionConfig `mapstructure:"fraud_detection"`
	Tax            TaxConfig            `mapstructure:"tax"`
	LoyaltyPoints  LoyaltyPointsConfig  `mapstructure:"loyalty_points"`
}

type DiscountConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	MaxPercentage  float64 `mapstructure:"max_percentage"`
	MaxFixedAmount float64 `mapstructure:"max_fixed_amount"`
}

type CashbackConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	Tier1Threshold  float64 `mapstructure:"tier1_threshold"`
	Tier1Percentage float64 `mapstructure:"tier1_percentage"`
	Tier2Percentage float64 `mapstructure:"tier2_percentage"`
}

type FraudDetectionConfig struct {
	Enabled                  bool          `mapstructure:"enabled"`
	MaxRiskScore             int           `mapstructure:"max_risk_score"`
	VelocityCheckWindow      time.Duration `mapstructure:"velocity_check_window"`
	MaxTransactionsPerWindow int           `mapstructure:"max_transactions_per_window"`
}

type TaxConfig struct {
	Enabled     bool               `mapstructure:"enabled"`
	DefaultRate float64            `mapstructure:"default_rate"`
	Rates       map[string]float64 `mapstructure:"rates"`
}

type LoyaltyPointsConfig struct {
	Enabled                 bool    `mapstructure:"enabled"`
	PointsToCurrencyRatio   float64 `mapstructure:"points_to_currency_ratio"`
	MaxRedemptionPercentage float64 `mapstructure:"max_redemption_percentage"`
}

type NotificationsConfig struct {
	Email   EmailConfig   `mapstructure:"email"`
	SMS     SMSConfig     `mapstructure:"sms"`
	Webhook WebhookConfig `mapstructure:"webhook"`
	Audit   AuditConfig   `mapstructure:"audit"`
}

type EmailConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	SMTPHost       string `mapstructure:"smtp_host"`
	SMTPPort       int    `mapstructure:"smtp_port"`
	FromAddress    string `mapstructure:"from_address"`
	WorkerPoolSize int    `mapstructure:"worker_pool_size"`
}

type SMSConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Provider  string `mapstructure:"provider"`
	RateLimit int    `mapstructure:"rate_limit"`
}

type WebhookConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}

type AuditConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	LogPath string `mapstructure:"log_path"`
}

type MetricsConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	ExportInterval time.Duration `mapstructure:"export_interval"`
}

type CLIConfig struct {
	PageSize int           `mapstructure:"page_size"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Theme    string        `mapstructure:"theme"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	v.SetEnvPrefix("ECOMMERCE")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "E-Commerce Payment System")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("database.driver", "sqlite3")
	v.SetDefault("database.path", "data/ecommerce.db")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("payment.timeout", "30s")
	v.SetDefault("payment.retry_attempts", 3)
}
