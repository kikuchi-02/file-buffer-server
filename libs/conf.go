package libs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	Port     int
	Endpoint string
}

type DBSettings struct {
	Name     string `yaml:"DATABASE_NAME"`
	User     string `yaml:"DATABASE_USER"`
	Host     string `yaml:"DATABASE_HOST"`
	Port     int    `yaml:"DATABASE_PORT"`
	Password string `yaml:"DATABASE_PASSWORD"`
}

const (
	MaxArrayLength   = 1e3 * 5
	MaxPassedMinutes = 30
	// more than 2
	Concurrency = 5
)

func LoadDBSettings(filepath string) {
	var settings DBSettings
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(bytes, &settings)
	port := strconv.Itoa(settings.Port)

	os.Setenv("DB_NAME", settings.Name)
	os.Setenv("DB_USER", settings.User)
	os.Setenv("DB_HOST", settings.Host)
	os.Setenv("DB_PORT", port)
	os.Setenv("DB_PASSWORD", settings.Password)
}

func LoadSettings(filepath string) Settings {
	var m map[string]interface{}

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(bytes, &m)
	settings := Settings{Port: m["PORT"].(int), Endpoint: m["ENDPOINT"].(string)}
	fmt.Println("Settings:", settings)
	return settings
}
