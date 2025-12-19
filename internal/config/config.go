package config

import (
	"log"
	"os"
	"time"
)

const (
	TaskManagerAddr         = "TASK_MANAGER_ADDR"
	TaskManagerPort         = ":8080"
	TaskManagerSqlitePath   = "TASK_MANAGER_SQLITE_PATH"
	TaskManagerPollInterval = 15 * time.Second
	TaskManagerSqliteDB     = "tasks.db"
)

type Config struct {
	Addr        string
	SQLitePath  string
	ReadTimeout time.Duration
}

func getenv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

func Load() Config {
	addr := getenv(TaskManagerAddr, TaskManagerPort)
	dbPath := getenv(TaskManagerSqlitePath, TaskManagerSqliteDB)

	readTimeout := TaskManagerPollInterval

	log.Printf("using addr=%s sqlite_path=%s", addr, dbPath)

	return Config{
		Addr:        addr,
		SQLitePath:  dbPath,
		ReadTimeout: readTimeout,
	}
}
