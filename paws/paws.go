package paws

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

const (
	TABLE_INFO_ROW  = iota // 0
	IMPORT_INFO_ROW        // 1
	KEY_INFO_ROW           // 2
	ALIAS_INFO_ROW         // 3
	FIELD_INFO_ROW         // 4
	START_DATA_ROW         // 5
)

type Column struct {
	Name, Alias string
	Key, Import bool
	Index       int
}

type Data struct {
	Value string
	Index int
}

func GetPAWSInfo(envString, defaultValue string) string {
	godotenv.Load(".env")
	envValue := os.Getenv(envString)
	fmt.Println(envString, envValue)
	if envValue != "" {
		return envValue
	}

	return defaultValue

}
