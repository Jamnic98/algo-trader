package setup

import "os"

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

    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        dsn = "host=host.docker.internal user=postgres password=dbpassword123 dbname=algo_trader_dev port=5432 sslmode=disable"
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    return Config{
        Env:  env,
        Dsn:  dsn,
        Port: port,
    }
}
