package entity

import (
	"github.com/mitchellh/mapstructure"
	. "logDog/common"
	"path/filepath"
	"regexp"
)

type Input struct {
	Type          string
	Path          string
	Files         []string
	Format        string
	Regex         string
	CompiledRegex *regexp.Regexp
}

func NewInputs(inputData map[string]interface{}) (map[string]Input, error) {
	inputs := make(map[string]Input)
	for label, inputObj := range inputData {
		input := Input{}
		err := mapstructure.Decode(inputObj, &input)
		if err != nil {
			return nil, err
		}
		if input.Regex != "" {
			input.CompiledRegex, err = regexp.Compile(input.Regex)
			if err != nil {
				return nil, err
			}
		} else {
			input.CompiledRegex = nil
		}
		input.Files, err = filepath.Glob(input.Path)
		if err != nil {
			return nil, err
		}
		Logger.Info(input.Files)
		inputs[label] = input
	}
	return inputs, nil
}
