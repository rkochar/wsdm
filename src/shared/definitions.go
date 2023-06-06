package shared

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type ServiceName string

const (
	StockService      ServiceName = "stock"
	OrderService      ServiceName = "order"
	PaymentService    ServiceName = "payment"
	APIGatewayService ServiceName = "api-gateway"
	LockmasterService ServiceName = "lockmaster"
)

type Config struct {
	StockNum      int64 `yaml:"stockdb"`
	OrderNum      int64 `yaml:"orderdb"`
	PaymentNum    int64 `yaml:"paymentdb"`
	APINum        int64 `yaml:"api-gateway"`
	LockmasterNum int64 `yaml:"lockmasterdb"`
}

func GetNumOfServices(serviceName ServiceName) (error, *int64) {
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

	serviceFields := map[ServiceName]*int64{
		StockService:      &config.StockNum,
		OrderService:      &config.OrderNum,
		PaymentService:    &config.PaymentNum,
		APIGatewayService: &config.APINum,
		LockmasterService: &config.LockmasterNum,
	}
	if fieldPtr, ok := serviceFields[serviceName]; ok {
		return nil, fieldPtr
	}
	return errors.New(fmt.Sprintf("Unsupported service name: %s", serviceName)), nil
}
