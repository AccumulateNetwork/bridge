package config

import (
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/jinzhu/configor"
	"github.com/mcuadros/go-defaults"
)

// App config struct
type Config struct {
	App struct {
		APIPort  int `required:"true" default:"13311" json:"apiPort" form:"apiPort" query:"apiPort"`
		LogLevel int `required:"true" default:"4" json:"logLevel" form:"logLevel" query:"logLevel"`
	}
	ACME struct {
		Node       string `required:"true" default:"" json:"node" form:"node" query:"node"`
		BridgeADI  string `required:"true" default:"" json:"bridgeADI" form:"bridgeADI" query:"bridgeADI"`
		KeyBook    string `required:"true" default:"book" json:"keyBook" form:"keyBook" query:"keyBook"`
		PrivateKey string `required:"true" default:"" json:"privateKey" form:"privateKey" query:"privateKey"`
	}
	EVM struct {
		Node           string  `required:"false" default:"" json:"node" form:"node" query:"node"`
		ChainId        int     `required:"true" default:"1" json:"chainId" form:"chainId" query:"chainId"`
		SafeAddress    string  `required:"true" default:"" json:"safeAddress" form:"safeAddress" query:"safeAddress"`
		BridgeAddress  string  `required:"true" default:"" json:"bridgeAddress" form:"bridgeAddress" query:"bridgeAddress"`
		PrivateKey     string  `required:"true" default:"" json:"privateKey" form:"privateKey" query:"privateKey"`
		MaxGasFee      float64 `required:"true" default:"30" json:"maxGasFee" form:"maxGasFee" query:"maxGasFee"`
		MaxPriorityFee float64 `required:"true" default:"2" json:"maxPriorityFee" form:"maxPriorityFee" query:"maxPriorityFee"`
	}
}

// Create config from configFile
func NewConfig(configFile string) (*Config, error) {

	config := new(Config)
	defaults.SetDefaults(config)

	configBytes, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			return nil, err
		}
	}

	if err := configor.Load(config); err != nil {
		return nil, err
	}
	return config, nil
}

func UpdateConfig(configFile string, newConf *Config) error {

	newYaml, err := yaml.Marshal(&newConf)
	if err != nil {
		return err
	}

	// write to file
	f, err := os.Create("/tmp/newconfig")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFile, newYaml, 0644)
	if err != nil {
		return err
	}

	f.Close()

	return nil

}
