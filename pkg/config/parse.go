package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ReadConfigurationFile tries to read the passed configuration file and parse it into a NovelConfig struct
func ReadConfigurationFile(fileName string) (novelConfig *NovelConfig, err error) {
	content, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(content, &novelConfig)
	if err != nil {
		return nil, err
	}
	// set base directory for includes and the like
	baseDirectory := filepath.Dir(fileName)
	novelConfig.BaseDirectory, err = filepath.Abs(baseDirectory)
	return novelConfig, err
}
