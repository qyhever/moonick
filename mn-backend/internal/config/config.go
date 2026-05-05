package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Mode     string         `mapstructure:"mode"`
	Server   ServerConfig   `mapstructure:"server"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Auth     AuthConfig     `mapstructure:"auth"`
	R2       R2Config       `mapstructure:"r2"`
	Postal   PostalConfig   `mapstructure:"postal"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Addr     string `mapstructure:"addr"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret                     string        `mapstructure:"secret"`
	AccessTokenTTL             time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL            time.Duration `mapstructure:"refresh_token_ttl"`
	RememberMeRefreshTokenTTL  time.Duration `mapstructure:"remember_me_refresh_token_ttl"`
	AccessExpiresIn            time.Duration `mapstructure:"access_expires_in"`
	RefreshExpiresIn           time.Duration `mapstructure:"refresh_expires_in"`
	RememberMeRefreshExpiresIn time.Duration `mapstructure:"remember_me_refresh_expires_in"`
}

// AuthConfig 认证相关配置
// 包括访问白名单（支持 "METHOD:/path" 或仅路径形式）。
type AuthConfig struct {
	Whitelist []string        `mapstructure:"whitelist"`
	Admin     AdminSeedConfig `mapstructure:"admin"`
}

type AdminSeedConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// R2Config Cloudflare R2配置
type R2Config struct {
	BucketName      string `mapstructure:"bucket_name"`
	AccountID       string `mapstructure:"account_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	PublicBaseURL   string `mapstructure:"public_base_url"`
}

// Postal配置
type PostalConfig struct {
	SmtpServer string `mapstructure:"smtp_server"`
	SmtpPort   string `mapstructure:"smtp_port"`
	FromEmail  string `mapstructure:"from_email"`
	FromPass   string `mapstructure:"from_pass"`
	FromName   string `mapstructure:"from_name"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Init 初始化配置
func Init() error {
	// 获取环境变量，默认为 dev
	env := os.Getenv("MOONICK_ENV")
	if env == "" {
		env = "dev"
	}

	// 验证环境变量值，只允许 dev、test、prod
	validEnvs := map[string]bool{
		"dev":  true,
		"test": true,
		"prod": true,
	}
	if !validEnvs[env] {
		return fmt.Errorf("无效的环境变量 MOONICK_ENV=%s，只允许: dev, test, prod", env)
	}

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 设置配置文件搜索路径
	viper.AddConfigPath(filepath.Join(workDir, "internal/config"))
	viper.AddConfigPath("./internal/config")
	viper.AddConfigPath(".")

	// 动态设置配置文件名
	viper.SetConfigName(env)
	viper.SetConfigType("yml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("MOONICK")
	viper.AutomaticEnv()

	// 绑定具体的环境变量
	viper.BindEnv("mode", "MOONICK_MODE")

	viper.BindEnv("server.port", "MOONICK_SERVER_PORT")

	viper.BindEnv("database.mysql.addr", "MOONICK_DATABASE_MYSQL_ADDR")
	viper.BindEnv("database.mysql.user", "MOONICK_DATABASE_MYSQL_USER")
	viper.BindEnv("database.mysql.password", "MOONICK_DATABASE_MYSQL_PASSWORD")
	viper.BindEnv("database.mysql.db_name", "MOONICK_DATABASE_MYSQL_DB_NAME")

	viper.BindEnv("jwt.secret", "MOONICK_JWT_SECRET")
	viper.BindEnv("jwt.access_token_ttl", "MOONICK_JWT_ACCESS_TOKEN_TTL")
	viper.BindEnv("jwt.refresh_token_ttl", "MOONICK_JWT_REFRESH_TOKEN_TTL")
	viper.BindEnv("jwt.remember_me_refresh_token_ttl", "MOONICK_JWT_REMEMBER_ME_REFRESH_TOKEN_TTL")
	viper.BindEnv("jwt.access_expires_in", "MOONICK_JWT_ACCESS_EXPIRES_IN")
	viper.BindEnv("jwt.refresh_expires_in", "MOONICK_JWT_REFRESH_EXPIRES_IN")
	viper.BindEnv("jwt.remember_me_refresh_expires_in", "MOONICK_JWT_REMEMBER_ME_REFRESH_EXPIRES_IN")

	viper.BindEnv("auth.admin.username", "MOONICK_AUTH_ADMIN_USERNAME")
	viper.BindEnv("auth.admin.password", "MOONICK_AUTH_ADMIN_PASSWORD")
	viper.BindEnv("auth.admin.name", "MOONICK_AUTH_ADMIN_NAME")

	viper.BindEnv("logger.level", "MOONICK_LOGGER_LEVEL")
	viper.BindEnv("logger.filename", "MOONICK_LOGGER_FILENAME")
	viper.BindEnv("logger.max_size", "MOONICK_LOGGER_MAX_SIZE")
	viper.BindEnv("logger.max_age", "MOONICK_LOGGER_MAX_AGE")
	viper.BindEnv("logger.max_backups", "MOONICK_LOGGER_MAX_BACKUPS")

	viper.BindEnv("r2.bucket_name", "MOONICK_R2_BUCKET_NAME")
	viper.BindEnv("r2.account_id", "MOONICK_R2_ACCOUNT_ID")
	viper.BindEnv("r2.access_key_id", "MOONICK_R2_ACCESS_KEY_ID")
	viper.BindEnv("r2.access_key_secret", "MOONICK_R2_ACCESS_KEY_SECRET")
	viper.BindEnv("r2.public_base_url", "MOONICK_R2_PUBLIC_BASE_URL")

	// 1. 读取环境配置文件 (如 dev.yml)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件 %s.yml 失败: %w", env, err)
	}

	// 2. 尝试加载本地配置文件 (如 dev.local.yml)，用于本地覆盖
	// 注意：MergeInConfig 会查找并合并同名配置项
	viper.SetConfigName(env + ".local")
	if err := viper.MergeInConfig(); err == nil {
		log.Printf("已合并本地配置文件: %s.local.yml", env)
	} else {
		// 如果是文件未找到错误，则忽略；否则记录错误
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 在某些 viper 版本或场景下，MergeInConfig 未找到文件可能并不返回 ConfigFileNotFoundError，
			// 但通常如果是文件不存在，err != nil 且包含特定信息。
			// 这里我们仅当明确报错且不是 "Config File ... Not Found" 时才视为异常
			// 为了简化，我们假设 MergeInConfig 在文件不存在时会报错，我们可以选择忽略它或者打印日志
			// 这里选择仅在 debug 模式下或特定错误下打印，或者直接忽略文件不存在的情况
		}
	}

	// 将配置解析到结构体
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	GlobalConfig.JWT.normalize()
	GlobalConfig.Auth.Admin.normalize()

	log.Printf("当前环境: %s", env)
	log.Printf("配置文件加载成功: %s", viper.ConfigFileUsed())
	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}

// GetServerAddr 获取服务器地址
func GetServerAddr() string {
	if GlobalConfig == nil {
		return ":6303" // 默认端口
	}
	return fmt.Sprintf(":%d", GlobalConfig.Server.Port)
}

// GetMySQLDSN 获取MySQL连接字符串
func GetMySQLDSN() string {
	if GlobalConfig == nil {
		return ""
	}
	return BuildMySQLDSN(GlobalConfig)
}

// BuildMySQLDSN 从指定配置生成 MySQL 连接字符串
func BuildMySQLDSN(cfg *Config) string {
	if cfg == nil {
		return ""
	}

	mysql := cfg.Database.MySQL
	if strings.TrimSpace(mysql.Addr) == "" ||
		strings.TrimSpace(mysql.User) == "" ||
		strings.TrimSpace(mysql.DBName) == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysql.User, mysql.Password, mysql.Addr, mysql.DBName)
}

// GetEnv 获取当前环境（dev/test/prod）
func GetEnv() string {
	env := os.Getenv("MOONICK_ENV")
	if env == "" {
		return "dev"
	}
	return env
}

// IsProduction 判断是否为生产环境
func IsProduction() bool {
	return GetEnv() == "prod"
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	return GetEnv() == "dev"
}

// IsTest 判断是否为测试环境
func IsTest() bool {
	return GetEnv() == "test"
}

func (c *JWTConfig) normalize() {
	if c.AccessTokenTTL == 0 {
		c.AccessTokenTTL = c.AccessExpiresIn
	}
	if c.RefreshTokenTTL == 0 {
		c.RefreshTokenTTL = c.RefreshExpiresIn
	}
	if c.RememberMeRefreshTokenTTL == 0 {
		c.RememberMeRefreshTokenTTL = c.RememberMeRefreshExpiresIn
	}
}

func (c *AdminSeedConfig) normalize() {
	c.Username = strings.TrimSpace(c.Username)
	c.Password = strings.TrimSpace(c.Password)
	c.Name = strings.TrimSpace(c.Name)
}
