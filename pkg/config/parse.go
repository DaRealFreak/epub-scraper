package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// ReadConfigurationFile tries to read the passed configuration file and parse it into a NovelConfig struct
func ReadConfigurationFile(fileName string) (novelConfig *NovelConfig, err error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(content, &novelConfig)
	return novelConfig, err
}
