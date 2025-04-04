package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// AppConfig 顶层配置结构，匹配yaml文件中的app键
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// Config 应用配置结构
type Config struct {
	App AppConfig `mapstructure:"app"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Timeout      time.Duration `mapstructure:"timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"` // 匹配环境变量
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LogConfig struct {
	Level   string `mapstructure:"level"`
	File    string `mapstructure:"file"`    // 日志文件路径
	Console bool   `mapstructure:"console"` // 是否同时输出到控制台
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	AccessTokenExp    time.Duration `mapstructure:"access_token_exp"`
	RefreshTokenExp   time.Duration `mapstructure:"refresh_token_exp"`
	Issuer            string        `mapstructure:"issuer"`
}

// LoadConfig 加载配置
func LoadConfig(path string) (*AppConfig, error) {
	// 初始化 viper
	viper.SetConfigFile(path)

	// 设置环境变量前缀和分隔符
	viper.SetEnvPrefix("app")

	// 启用环境变量支持
	viper.AutomaticEnv()

	// 先读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 设置默认值
	setDefaults(&config.App)

	return &config.App, nil
}

// 设置默认值
func setDefaults(config *AppConfig) {
	// 服务器默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}

	// JWT默认值
	if config.JWT.AccessTokenExp == 0 {
		config.JWT.AccessTokenExp = 24 * time.Hour
	}
	if config.JWT.RefreshTokenExp == 0 {
		config.JWT.RefreshTokenExp = 7 * 24 * time.Hour
	}
	if config.JWT.Issuer == "" {
		config.JWT.Issuer = "go-rest-starter"
	}
	// 必须有Secret，否则不安全
	if config.JWT.Secret == "" {
		config.JWT.Secret = "default-secret-key-please-change-in-production"
	}
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSLMode)
}
