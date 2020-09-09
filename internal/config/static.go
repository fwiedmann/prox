package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

var (
	ErrorInvalidFileType             = errors.New("given file type is invalid, only .yaml or yml is allowed")
	ErrorDuplicatedPortConfiguration = errors.New("static configuration has an invalid duplicated port configuration")
)

// Static
type Static struct {
	Ports            []Port `yaml:"ports"`
	Cache            Cache  `yaml:"cache"`
	AccessLogEnabled bool   `yaml:"access-log-enabled"`
	InfraPort        uint16 `yaml:"infra-port"`
}

// Port
type Port struct {
	Name       string `yaml:"name"`
	Addr       uint16 `yaml:"port"`
	TlSEnabled bool   `yaml:"tls"`
}

// Cache
type Cache struct {
	Enabled                bool  `yaml:"enabled"`
	CacheMaxSizeInMegaByte int64 `yaml:"cache-max-size-in-mega-byte"`
}

// ParseStaticFile
func ParseStaticFile(path string) (Static, error) {
	file, err := os.Open(path)
	if err != nil {
		return Static{}, err
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return Static{}, err
	}

	if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yaml") {
		return Static{}, ErrorInvalidFileType
	}

	var config Static
	if err := yaml.Unmarshal(content, &config); err != nil {
		return Static{}, err
	}
	log.Debugf("Parsed static config file \"%s\": %+v", path, config)

	if config.InfraPort == 0 {
		config.InfraPort = 9100
	}

	if hasDuplicates(config.Ports, config.InfraPort) {
		return Static{}, ErrorDuplicatedPortConfiguration
	}
	return config, nil
}

func hasDuplicates(ports []Port, infraPort uint16) bool {
	var hasDuplicatePortsAddr bool
	var hasDuplicatesNames bool
	var hasDuplicatesWithInfraPort bool
	duplicatesAddr := make(map[string]int)
	duplicatesNames := make(map[string]int)

	for _, p1 := range ports {
		countPortAddr := 0
		countNames := 0

		if p1.Addr == infraPort {
			hasDuplicatesWithInfraPort = true
			log.Errorf("static port configuration \"%s\" has the same port as the infra port on %d", p1.Name, infraPort)
		}
		for _, p2 := range ports {
			if p1.Addr == p2.Addr {
				countPortAddr++
			}
			if p1.Name == p2.Name {
				countNames++
			}
		}
		if countPortAddr > 1 {
			hasDuplicatePortsAddr = true
			duplicatesAddr[(p1.Name)] = countPortAddr
		}
		if countNames > 1 {
			hasDuplicatesNames = true
			duplicatesNames[(p1.Name)] = countNames
		}

	}
	if hasDuplicatePortsAddr {
		log.Error("static port configuration has duplicated port addresses:")
		for key, val := range duplicatesAddr {
			log.Errorf("port address: \"%s\", count: %d", key, val)
		}
	}

	if hasDuplicatesNames {
		log.Error("static port configuration has duplicated port names:")
		for key, val := range duplicatesAddr {
			log.Errorf("port-name: \"%s\", count: %d", key, val)
		}
	}

	return hasDuplicatePortsAddr || hasDuplicatesNames || hasDuplicatesWithInfraPort
}
