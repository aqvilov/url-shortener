package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// в конфиге храним переменную окружения (local, prod, test etc.)
type Config struct {
	Env         string     `yaml:"env"`
	StoragePath string     `yaml:"storage_path"`
	HTTPServer  HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Addr        string        `yaml:"addr"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

// загружаем конфиг
func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Предупреждение: файл .env не найден: %v", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./local/local.yaml"
		log.Printf("use default config path: %s", configPath)
	}
	// проверка сущуствования файла через os.Stat
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Panic("config file does not exist")
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Panic("couldnt read config")
	}

	return config
}
