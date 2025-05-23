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
}

func New() *Config {
	grpcPort := flag.Int("grpc-port", 50051, "The gRPC server port")
	restPort := flag.Int("rest-port", 8080, "The REST server port")
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

	return &Config{
		DBUser:                  DBUser,
		DBPassword:              DBPassword,
		DBName:                  DBName,
		DBHost:                  DBHost,
		DBPort:                  DBPortString,
		DBItemLimit:             10, //TODO: Add an option to configure this or find a smarter way to set this
		EventEmitterBufferLimit: 10, //TODO: Add an option to configure this or find a smarter way to set this
		GRPCPort:                *grpcPort,
		RESTPort:                *restPort,
		AuthToken:               authToken,
	}
}

func (c *Config) GetDBURI() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
