package entity

import (
	"github.com/mitchellh/mapstructure"
	. "logDog/common"
	"regexp"
)

type Input struct {
	Type          string
	Path          string
	Format        string
	Regex         string
	CompiledRegex *regexp.Regexp
}

func NewInputs(inputData map[string]interface{}) map[string]Input {
	inputs := make(map[string]Input)
	for label, inputObj := range inputData {
		input := Input{}
		err := mapstructure.Decode(inputObj, &input)
		if err != nil {
			Logger.Error(err)
			return nil
		}
		if input.Regex != "" {
			input.CompiledRegex, err = regexp.Compile(input.Regex)
			if err != nil {
				return nil
			}
		} else {
			input.CompiledRegex = nil
		}
		inputs[label] = input
	}
	return inputs
}
