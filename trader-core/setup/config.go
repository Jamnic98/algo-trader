package setup

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    Env  string
    Dsn  string
    Port string
}

func GetConfig() Config {
    env := os.Getenv("ENV")
    if env == "" {
        env = "local"
    }

    fmt.Println("ENV:", env)

    // Try to load .env file, but don’t fail if it doesn’t exist
    envFile := fmt.Sprintf(".env.%s", env)
    if err := godotenv.Load(envFile); err != nil {
        log.Println("No env file found, using system env:", envFile)
    } else {
        fmt.Println("Loaded env file:", envFile)
    }

    dsn := os.Getenv("DB_DSN")
    port := os.Getenv("TRADER_CORE_PORT")
    if port == "" {
        port = "8080"
    }

    return Config{
        Env:  env,
        Dsn:  dsn,
        Port: port,
    }
}
