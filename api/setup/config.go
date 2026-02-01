package setup

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
  Env string
  Dsn string
  Port string
}

func GetConfig() Config {
  env := os.Getenv("ENV")
  if env == "" {
    env = "local"
  }

  envFile := fmt.Sprintf(".env.%s", env)
  if err := godotenv.Load(envFile); err != nil {
    log.Fatalf("Error loading %s file: %v", envFile, err)
  }

  dsn := os.Getenv("DB_DSN")
  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  return Config{
    Env: env,
    Dsn: dsn,
    Port: port,
  }
}
