package conf

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

func ReadCommands(fileName string) (commands map[string]Command, err error) {
	buf, err := os.ReadFile(fileName)
	if err != nil {
		return
	}

	err = json.Unmarshal(buf, &commands)
	return
}

func ReadConfig(fileName string) (conf Configuration, err error) {
	buf, err := os.ReadFile(fileName)
	if err != nil {
		return Configuration{}, err
	}

	err = yaml.Unmarshal(buf, &conf)
	return
}
