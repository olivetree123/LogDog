package entity

import (
	"github.com/mitchellh/mapstructure"
	. "logDog/common"
)

type Output struct {
	Type        string
	HttpURL     string            `mapstructure:"http_url"`
	HttpHeaders map[string]string `mapstructure:"http_headers"`
	RedisAddr   string            `mapstructure:"redis_addr"`
	RedisDB     int               `mapstructure:"redis_db"`
	RedisKey    string            `mapstructure:"redis_key"`
}

func NewOutputs(outputData map[string]interface{}) map[string]Output {
	outputs := make(map[string]Output)
	for label, outputObj := range outputData {
		output := Output{}
		err := mapstructure.Decode(outputObj, &output)
		if err != nil {
			Logger.Error(err)
			return nil
		}
		outputs[label] = output
	}
	return outputs
}
