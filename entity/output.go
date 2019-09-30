package entity

import (
	"github.com/mitchellh/mapstructure"
	"os"
)

type Output struct {
	Type          string
	HttpURL       string            `mapstructure:"http_url"`
	HttpHeaders   map[string]string `mapstructure:"http_headers"`
	RedisAddr     string            `mapstructure:"redis_addr"`
	RedisDB       int               `mapstructure:"redis_db"`
	RedisKey      string            `mapstructure:"redis_key"`
	MySQLHost     string            `mapstructure:"mysql_host"`
	MySQLPort     int               `mapstructure:"mysql_port"`
	MySQLDb       string            `mapstructure:"mysql_db"`
	MySQLUser     string            `mapstructure:"mysql_user"`
	MySQLPassword string            `mapstructure:"mysql_password"`
	HostName      string
}

func NewOutputs(outputData map[string]interface{}) (map[string]Output, error) {
	outputs := make(map[string]Output)
	for label, outputObj := range outputData {
		output := Output{}
		err := mapstructure.Decode(outputObj, &output)
		if err != nil {
			return nil, err
		}
		output.HostName, err = os.Hostname()
		if err != nil {
			return nil, err
		}
		outputs[label] = output
	}
	return outputs, nil
}
