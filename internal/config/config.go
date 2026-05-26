package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// в конфиге храним переменную окружения (local, prod, test etc.),  путь к бд, хотя по сути он не обязательный тут
type Config struct {
	Env         string `yaml:"env"`
	StoragePath string `yaml:"storage_path"`
}

type HTTPServer struct {
	Addr        string        `yaml:"addr"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

// загружаем конфиг
func Load() Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "../../local/local.yaml" // конфиг отсюда
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
