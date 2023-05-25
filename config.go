package service

import (
	"encoding/json"
	"fmt"
	"os"
)

type GoogleConfig struct {
	Project string
	LogName string
}

type BaseConfig struct {
	ServerType ServerType
	ConfigName string
	Version    string
	Google     GoogleConfig
	SentryDsn  *string
}

type IBaseConfig interface {
	GetVersionString() string
	SetServerType(string)
	GetServerType() ServerType
	IsProductionServer() bool
}

func (c *BaseConfig) GetVersionString() string {
	if c == nil {
		return ""
	} else {
		return c.Version + "-" + c.ConfigName
	}
}

func (c *BaseConfig) GetServerType() ServerType {
	if c == nil {
		return Development
	} else {
		return c.ServerType
	}
}

func (c *BaseConfig) GetSentryDsn() *string {
	if c == nil {
		return nil
	}
	return c.SentryDsn
}

func (c *BaseConfig) IsProductionServer() bool {
	switch c.ServerType {
	case Production, LiveTest:
		return true
	}

	return false
}

func (c *BaseConfig) SetServerType(envServerType string) {
	if c == nil {
		return
	}

	switch envServerType {
	case "production", "prod":
		c.ServerType = Production
	case "staging", "stage":
		c.ServerType = Staging
	case "livetest":
		c.ServerType = LiveTest
	case "uat":
		c.ServerType = UAT
	case "local":
		c.ServerType = Local
	default:
		c.ServerType = Development
	}
}

func ReadConfig(config interface{}, envServerType string, configPathBuilder func(string) string) error {
	baseConfig, isBaseConfig := config.(IBaseConfig)
	if isBaseConfig {
		baseConfig.SetServerType(envServerType)
	}

	configPath := configPathBuilder("config.json")

	var mergedConfig = make(map[string]interface{})
	var overlayConfig = make(map[string]interface{})

	//  Read base config for merging
	if err := readConfigPath(configPath, &mergedConfig); err != nil {
		return err
	}

	//  Merge specific configuration if applicable
	var variantConfigPaths []string

	switch baseConfig.GetServerType() {
	case Staging:
		variantConfigPaths = []string{"config-stage.json", "config-staging.json"}
	case Production:
		variantConfigPaths = []string{"config-pro.json", "config-prod.json", "config-production.json"}
	case Development:
		variantConfigPaths = []string{"config-dev.json", "config-development.json"}
	case UAT:
		variantConfigPaths = []string{"config-uat.json"}
	case LiveTest:
		variantConfigPaths = []string{"config-livetest.json"}
	case Local:
		variantConfigPaths = []string{"config-local.json"}
	}

	var configReadError error
	for _, variantConfigFile := range variantConfigPaths {
		variantConfigPath := configPathBuilder(variantConfigFile)

		if _, err := os.Stat(variantConfigPath); os.IsNotExist(err) {
			//  Skip if missing
			continue
		}

		if configReadError = readConfigPath(variantConfigPath, &overlayConfig); configReadError == nil {
			break
		}
	}
	if configReadError != nil {
		return configReadError
	}

	//  Merge an overlay configuration with the base configuration if present
	if len(overlayConfig) > 0 {
		for key, value := range overlayConfig {
			if baseValue, ok := mergedConfig[key]; !ok {
				//  Missing base key, add it
				mergedConfig[key] = value
			} else if baseMap, isMapBase := baseValue.(map[string]interface{}); !isMapBase {
				//  Base config is not a map, overwrite with overlay value
				mergedConfig[key] = value
			} else if overlayMap, isMapOverlay := value.(map[string]interface{}); isMapBase && isMapOverlay {
				//  Merge base and overlay map
				for subKey, subValue := range overlayMap {
					baseMap[subKey] = subValue
				}
			} else {
				//  Base config is a map, overwrite with a non-map overlay value
				mergedConfig[key] = value
			}
		}
	}

	//  Read the merged configuration to config struct, requires translating back and forth from json
	var mergedJson []byte
	var err error
	if mergedJson, err = json.Marshal(mergedConfig); err == nil {
		err = json.Unmarshal(mergedJson, config)
	}

	return err
}

func readConfigPath(path string, config interface{}) error {
	file, err := os.Open(path)

	if err != nil {
		err = fmt.Errorf("ReadConfig path error: %v", err)
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)

	return err
}
