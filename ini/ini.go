package ini

import (
	"github.com/larspensjo/config"
	"errors"
)

func Get(file string, sec string) (map[string]string, error) {
	targetConfig := make(map[string]string)
	cfg, err := config.ReadDefault(file)
	if err != nil {
		return targetConfig, errors.New("unable to open config file or wrong fomart")
	}
	sections := cfg.Sections()
	if len(sections) == 0 {
		return targetConfig, errors.New("no " + sec + " config")
	}
	for _, section := range sections {
		if section != sec {
			continue
		}
		sectionData, _ := cfg.SectionOptions(section)
		for _, key := range sectionData {
			value, err := cfg.String(section, key)
			if err == nil {
				targetConfig[key] = value
			}
		}
		break
	}
	return targetConfig, nil
}

