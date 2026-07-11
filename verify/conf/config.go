package conf

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/he-end/verify-reverse/verify/log"
)

type Conf struct {
	*WAConf
	*EmailConf
	DBConf
	JWTConf
	AppEnv   string
	LogLevel string
}

type EmailConf struct {
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPHost string
}

type WAConf struct {
	TokenWhatsApp   string
	BaseURLGraphAPI string
	PhoneNumberID   string
	WhatsAppPhone   string
}

type DBConf struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func (d *DBConf) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.DBUser, d.DBPassword, d.DBHost, d.DBPort, d.DBName, d.DBSSLMode)
}

type JWTConf struct {
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
}

func GetEnv() *Conf {
	// workDir, _ := os.Getwd()
	if load := godotenv.Load(".env"); load != nil {
		log.Warn("env file not detected, using OS environment variables")
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}

	conf := Conf{
		AppEnv:   appEnv,
		LogLevel: logLevel,
		WAConf: &WAConf{
			TokenWhatsApp:   os.Getenv("TOKEN_WHATSAPP"),
			BaseURLGraphAPI: os.Getenv("BASE_URL_GRAPH_API"),
			PhoneNumberID:   os.Getenv("PHONE_NUMBER_ID"),
			WhatsAppPhone:   os.Getenv("WHATSAPP_PHONE"),
		},
		EmailConf: &EmailConf{
			SMTPPort: os.Getenv("SMTP_PORT"),
			SMTPHost: os.Getenv("SMTP_HOST"),
			SMTPUser: os.Getenv("SMTP_USER"),
			SMTPPass: os.Getenv("SMTP_PASS"),
		},
		DBConf: DBConf{
			DBHost:     os.Getenv("DB_HOST"),
			DBPort:     os.Getenv("DB_PORT"),
			DBUser:     os.Getenv("DB_USER"),
			DBPassword: os.Getenv("DB_PASSWORD"),
			DBName:     os.Getenv("DB_NAME"),
			DBSSLMode:  os.Getenv("DB_SSLMODE"),
		},
		JWTConf: JWTConf{
			JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
			JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
			JWTAccessTTL:     parseDuration(os.Getenv("JWT_ACCESS_TTL"), 15*time.Minute),
			JWTRefreshTTL:    parseDuration(os.Getenv("JWT_REFRESH_TTL"), 168*time.Hour),
		},
	}
	return &conf
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
