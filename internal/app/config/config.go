package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// 配置验证错误
var (
	ErrInvalidPort         = errors.New("invalid server port")
	ErrMissingDatabaseHost = errors.New("missing database host")
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
	Port         int           `mapstructure:"port" env:"SERVER_PORT"`
	Timeout      time.Duration `mapstructure:"timeout" env:"SERVER_TIMEOUT"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" env:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" env:"DB_DRIVER"`
	Host            string        `mapstructure:"host" env:"DB_HOST"`
	Port            int           `mapstructure:"port" env:"DB_PORT"`
	Username        string        `mapstructure:"username" env:"DB_USERNAME"`
	Password        string        `mapstructure:"password" env:"DB_PASSWORD"`
	DBName          string        `mapstructure:"dbname" env:"DB_NAME"`
	SSLMode         string        `mapstructure:"sslmode" env:"DB_SSLMODE"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host" env:"REDIS_HOST"`
	Port     int    `mapstructure:"port" env:"REDIS_PORT"`
	Password string `mapstructure:"password" env:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"db" env:"REDIS_DB"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string `mapstructure:"level" env:"LOG_LEVEL"`
	File    string `mapstructure:"file" env:"LOG_FILE"`
	Console bool   `mapstructure:"console" env:"LOG_CONSOLE"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string        `mapstructure:"secret" env:"JWT_SECRET"`
	AccessTokenExp  time.Duration `mapstructure:"access_token_exp" env:"JWT_ACCESS_TOKEN_EXP"`
	RefreshTokenExp time.Duration `mapstructure:"refresh_token_exp" env:"JWT_REFRESH_TOKEN_EXP"`
	Issuer          string        `mapstructure:"issuer" env:"JWT_ISSUER"`
}

// LoadConfig 加载配置
func LoadConfig(path string) (*AppConfig, error) {
	// 初始化 viper
	viper.SetConfigFile(path)

	// 设置环境变量前缀和分隔符
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定环境变量
	bindEnvVariables()

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

// 绑定环境变量
func bindEnvVariables() {
	// 服务器配置环境变量
	viper.BindEnv("app.server.port", "APP_SERVER_PORT")
	viper.BindEnv("app.server.timeout", "APP_SERVER_TIMEOUT")
	viper.BindEnv("app.server.read_timeout", "APP_SERVER_READ_TIMEOUT")
	viper.BindEnv("app.server.write_timeout", "APP_SERVER_WRITE_TIMEOUT")

	// 数据库配置环境变量
	viper.BindEnv("app.database.driver", "APP_DB_DRIVER")
	viper.BindEnv("app.database.host", "APP_DB_HOST")
	viper.BindEnv("app.database.port", "APP_DB_PORT")
	viper.BindEnv("app.database.username", "APP_DB_USERNAME")
	viper.BindEnv("app.database.password", "APP_DB_PASSWORD")
	viper.BindEnv("app.database.dbname", "APP_DB_NAME")
	viper.BindEnv("app.database.sslmode", "APP_DB_SSLMODE")
	viper.BindEnv("app.database.max_open_conns", "APP_DB_MAX_OPEN_CONNS")
	viper.BindEnv("app.database.max_idle_conns", "APP_DB_MAX_IDLE_CONNS")
	viper.BindEnv("app.database.conn_max_lifetime", "APP_DB_CONN_MAX_LIFETIME")

	// Redis配置环境变量
	viper.BindEnv("app.redis.host", "APP_REDIS_HOST")
	viper.BindEnv("app.redis.port", "APP_REDIS_PORT")
	viper.BindEnv("app.redis.password", "APP_REDIS_PASSWORD")
	viper.BindEnv("app.redis.db", "APP_REDIS_DB")

	// 日志配置环境变量
	viper.BindEnv("app.log.level", "APP_LOG_LEVEL")
	viper.BindEnv("app.log.file", "APP_LOG_FILE")
	viper.BindEnv("app.log.console", "APP_LOG_CONSOLE")

	// JWT配置环境变量
	viper.BindEnv("app.jwt.secret", "APP_JWT_SECRET")
	viper.BindEnv("app.jwt.access_token_exp", "APP_JWT_ACCESS_TOKEN_EXP")
	viper.BindEnv("app.jwt.refresh_token_exp", "APP_JWT_REFRESH_TOKEN_EXP")
	viper.BindEnv("app.jwt.issuer", "APP_JWT_ISSUER")
}

// 设置默认值
func setDefaults(config *AppConfig) {
	// 服务器默认值
	if config.Server.Port == 0 {
		config.Server.Port = 7001
	}
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30 * time.Second
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 15 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 15 * time.Second
	}

	// 数据库连接池默认值
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 20
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 1 * time.Hour
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
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	// 构建PostgreSQL DSN - 确保dbname参数正确
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.DBName, c.SSLMode)
	} else {
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSLMode)
	}
}
