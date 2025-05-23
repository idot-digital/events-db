package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	DBUser                  string
	DBPassword              string
	DBName                  string
	DBHost                  string
	DBPort                  string
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
}

func New() *Config {
	grpcPort := flag.Int("grpc-port", 50051, "The gRPC server port")
	restPort := flag.Int("rest-port", 8080, "The REST server port")
	clientBufferSize := flag.Int("client-buffer-size", 100, "Buffer size for client event channels")
	maxTotalClients := flag.Int("max-total-clients", 10000, "Maximum total number of clients across all subjects")
	streamBatchSize := flag.Int("stream-batch-size", 10, "Number of events to fetch in each stream batch")
	flag.Parse()

	DBUser, isSet := os.LookupEnv("MYSQL_USER")
	if !isSet {
		DBUser = "root"
	}
	DBPassword, isSet := os.LookupEnv("MYSQL_PASSWORD")
	if !isSet {
		DBPassword = "root"
	}
	DBName, isSet := os.LookupEnv("MYSQL_DATABASE_NAME")
	if !isSet {
		DBName = "root"
	}
	DBHost, isSet := os.LookupEnv("MYSQL_HOST")
	if !isSet {
		DBHost = "localhost"
	}
	DBPortString, isSet := os.LookupEnv("MYSQL_PORT")
	if !isSet {
		DBPortString = "3306"
	}
	authToken, isSet := os.LookupEnv("AUTH_TOKEN")
	if !isSet {
		authToken = "" // Empty token means no authentication required
	}
	tlsCertFile, _ := os.LookupEnv("TLS_CERT_FILE")
	tlsKeyFile, _ := os.LookupEnv("TLS_KEY_FILE")

	return &Config{
		DBUser:                  DBUser,
		DBPassword:              DBPassword,
		DBName:                  DBName,
		DBHost:                  DBHost,
		DBPort:                  DBPortString,
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
	}
}

func (c *Config) GetDBURI() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
