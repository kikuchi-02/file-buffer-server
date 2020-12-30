package libs

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	Port     int
	Endpoint string
}

func LoadSettings() Settings {
	var m map[string]interface{}

	bytes, err := ioutil.ReadFile("settings.yaml")
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(bytes, &m)
	settings := Settings{Port: m["PORT"].(int), Endpoint: m["ENDPOINT"].(string)}
	fmt.Println("Settings:", settings)
	return settings
}
