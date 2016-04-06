package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	ConnString string
	SourceData string
	ReadRows   string
	RunMode    string
}

func Load() (*Config, error) {
	godotenv.Load(".env")

	connString := os.Getenv("MSSQL_CONN_STRING")
	if connString == "" {
		return nil, fmt.Errorf("input MS SQL connection string")
	}

	sourceFile := os.Getenv("DATA_FILE")
	if sourceFile == "" {
		return nil, fmt.Errorf("input DATA_FILE")
	}

	runMode := os.Getenv("RUN_MODE")
	if runMode == "" {
		return nil, fmt.Errorf("input RUN_MODE")
	}

	readRows := os.Getenv("READ_ROWS")
	if readRows == "" {
		return nil, fmt.Errorf("input READ_ROWS")
	}

	return &Config{
		connString,
		sourceFile,
		readRows,
		runMode,
	}, nil
}
