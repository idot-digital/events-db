package config

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	DBUser                  string
	DBPassword              string
	DBName                  string
	DBHost                  string
	DBPort                  string
	DBTLS                   bool
	SkipSchemaCreation      bool
	DBItemLimit             int
	EventEmitterBufferLimit int
	GRPCPort                int
	RESTPort                int
	AuthToken               string
	TLSCertFile             string
	TLSKeyFile              string
	ClientBufferSize        int
	MaxTotalClients         int
	StreamBatchSize         int
	LogLevel                slog.Level
}

func New() *Config {
	grpcPort := flag.Int("grpc-port", 50051, "The gRPC server port")
	restPort := flag.Int("rest-port", 8080, "The REST server port")
	clientBufferSize := flag.Int("client-buffer-size", 100, "Buffer size for client event channels")
	maxTotalClients := flag.Int("max-total-clients", 10000, "Maximum total number of clients across all subjects")
	streamBatchSize := flag.Int("stream-batch-size", 10, "Number of events to fetch in each stream batch")
	flag.Parse()

	var DBUser, DBPassword, DBName, DBHost, DBPortString string
	var dbTLS bool

	// Check if MYSQL_URL is provided first
	if mysqlURL, isSet := os.LookupEnv("MYSQL_URL"); isSet {
		parsedURL, err := url.Parse(mysqlURL)
		if err != nil {
			// Fall back to individual env vars if URL parsing fails
			DBUser = getEnvWithDefault("MYSQL_USER", "root")
			DBPassword = getEnvWithDefault("MYSQL_PASSWORD", "root")
			DBName = getEnvWithDefault("MYSQL_DATABASE_NAME", "root")
			DBHost = getEnvWithDefault("MYSQL_HOST", "localhost")
			DBPortString = getEnvWithDefault("MYSQL_PORT", "3306")
			dbTLS = getEnvWithDefault("MYSQL_TLS", "false") == "true"
		} else {
			// Extract connection details from URL
			DBUser = parsedURL.User.Username()
			if password, hasPassword := parsedURL.User.Password(); hasPassword {
				DBPassword = password
			}
			DBHost = parsedURL.Hostname()
			if parsedURL.Port() != "" {
				DBPortString = parsedURL.Port()
			} else {
				DBPortString = "3306"
			}
			DBName = strings.TrimPrefix(parsedURL.Path, "/")
			
			// Check for SSL/TLS in query parameters
			queryParams := parsedURL.Query()
			if sslParam := queryParams.Get("ssl"); sslParam != "" {
				dbTLS = true
			}
			if tlsParam := queryParams.Get("tls"); tlsParam == "true" {
				dbTLS = true
			}
		}
	} else {
		// Use individual environment variables
		DBUser = getEnvWithDefault("MYSQL_USER", "root")
		DBPassword = getEnvWithDefault("MYSQL_PASSWORD", "root")
		DBName = getEnvWithDefault("MYSQL_DATABASE_NAME", "root")
		DBHost = getEnvWithDefault("MYSQL_HOST", "localhost")
		DBPortString = getEnvWithDefault("MYSQL_PORT", "3306")
		dbTLS = getEnvWithDefault("MYSQL_TLS", "false") == "true"
	}
	authToken, isSet := os.LookupEnv("AUTH_TOKEN")
	if !isSet {
		authToken = "" // Empty token means no authentication required
	}
	tlsCertFile, _ := os.LookupEnv("TLS_CERT_FILE")
	tlsKeyFile, _ := os.LookupEnv("TLS_KEY_FILE")
	
	skipSchemaCreation := getEnvWithDefault("SKIP_SCHEMA_CREATION", "false") == "true"

	logLevel := slog.LevelInfo
	levelStr := os.Getenv("LOG_LEVEL")
	switch levelStr {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return &Config{
		DBUser:                  DBUser,
		DBPassword:              DBPassword,
		DBName:                  DBName,
		DBHost:                  DBHost,
		DBPort:                  DBPortString,
		DBTLS:                   dbTLS,
		SkipSchemaCreation:      skipSchemaCreation,
		DBItemLimit:             10,
		EventEmitterBufferLimit: 100,
		GRPCPort:                *grpcPort,
		RESTPort:                *restPort,
		AuthToken:               authToken,
		TLSCertFile:             tlsCertFile,
		TLSKeyFile:              tlsKeyFile,
		ClientBufferSize:        *clientBufferSize,
		MaxTotalClients:         *maxTotalClients,
		StreamBatchSize:         *streamBatchSize,
		LogLevel:                logLevel,
	}
}

func (c *Config) GetDBURI() string {
	baseURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
	
	if c.DBTLS {
		baseURI += "&tls=true"
	}
	
	return baseURI
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, isSet := os.LookupEnv(key); isSet {
		return value
	}
	return defaultValue
}
