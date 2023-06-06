package shared

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const NUM_DBS = 3

type Config struct {
	StockNum      int64 `yaml:"stockdb"`
	OrderNum      int64 `yaml:"orderdb"`
	PaymentNum    int64 `yaml:"paymentdb"`
	APINum        int64 `yaml:"api-gateway"`
	LockmasterNum int64 `yaml:"lockmasterdb"`
}

func GetNumOfServices(serviceName string) (error, *int64) {
	fileName := "./config.yaml"

	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		return err, nil
	}

	var config Config
	unmarshalErr := yaml.Unmarshal(yamlFile, &config)
	if unmarshalErr != nil {
		fmt.Printf("Unmarshal error: %s\n", unmarshalErr)
		return unmarshalErr, nil
	}
	fmt.Printf("Config: %+v\n", config)

	serviceFields := map[string]*int64{
		"stock":       &config.StockNum,
		"order":       &config.OrderNum,
		"payment":     &config.PaymentNum,
		"api-gateway": &config.APINum,
		"lockmaster":  &config.LockmasterNum,
	}
	if fieldPtr, ok := serviceFields[serviceName]; ok {
		return nil, fieldPtr
	}
	return errors.New(fmt.Sprintf("Unsupported service name: %s", serviceName)), nil
}
