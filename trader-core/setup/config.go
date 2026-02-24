package setup

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey         string
	Env            string
	Dsn            string
	Port           string
	RedisURL       string
	TelegramChatID string
	TelegramKey    string
}

func GetConfig() Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	fmt.Println("APP_ENV:", env)

	// Try to load .env file, but don’t fail if it doesn’t exist
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		log.Println("No env file found, using system env:", envFile)
	} else {
		fmt.Println("Loaded env file:", envFile)
	}

	apiKey := os.Getenv("SERVER_API_KEY")
	dsn := os.Getenv("DB_DSN")
	port := os.Getenv("TRADER_CORE_PORT")
	telegramKey := os.Getenv("TELEGRAM_KEY")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	if port == "" {
		port = "8080"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/0"
	}

	return Config{
		ApiKey:         apiKey,
		Env:            env,
		Dsn:            dsn,
		Port:           port,
		RedisURL:       redisURL,
		TelegramChatID: telegramChatID,
		TelegramKey:    telegramKey,
	}
}
