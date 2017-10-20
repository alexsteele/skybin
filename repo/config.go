package repo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	core "skybin/core/proto"
)

type Config struct {
	UserId          string            `json:"userId"`
	NodeId          string            `json:"providerID"`
	DhtAddress      string            `json:"dhtAddress"`
	ProviderAddress string            `json:"providerAddress"`
	ApiAddress      string            `json:"apiAddress"`
	SeedAddresses   []string          `json:"seedAddresses"`
	LogFolder       string            `json:"logFolder"`
	LogEnabled      bool              `json:"logEnabled"`
	BlockSize       int               `json:"blockSize"`
	EncryptionType  string            `json:"encryptionType"`
	Redundancy      int               `json:"redundancy"`
	ProviderInfo    core.ProviderInfo `json:"providerInfo"`
}

type StorageOptions struct {
	FileName       string
	Redundancy     int
	BlockSize      int
	EncryptionType string
}

func (c *Config) DefaultStorageOpts(filename string) *StorageOptions {
	return &StorageOptions{
		FileName:       filename,
		Redundancy:     c.Redundancy,
		BlockSize:      c.BlockSize,
		EncryptionType: c.EncryptionType,
	}
}

func defaultConfig(userId string, nodeId string) *Config {
	return &Config{
		UserId:          userId,
		NodeId:          nodeId,
		DhtAddress:      "0.0.0.0:8001",
		ProviderAddress: "0.0.0.0:8002",
		ApiAddress:      "127.0.0.1:8003",
		LogFolder:       "",
		LogEnabled:      false,
		BlockSize:       1 << 20,
		EncryptionType:  "aes",
		Redundancy:      1,
		ProviderInfo: core.ProviderInfo{
			ID:           nodeId,
			MaxBlockSize: 1 << 30,
		},
	}
}

func loadConfig(filename string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Cannot read config: %s", err)
	}

	config := &Config{}
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		return nil, fmt.Errorf("Cannot not parse config: %s", err)
	}
	return config, nil
}
