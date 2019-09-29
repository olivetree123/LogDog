package entity

import (
	"github.com/mitchellh/mapstructure"
	"logDog/utils"
	"path/filepath"
	"regexp"
	"strings"
)

type Input struct {
	Type          string // file or docker
	Path          string
	CName         []string // 容器名称
	Files         []string
	Format        string
	Regex         string
	CompiledRegex *regexp.Regexp
	DockerLogPath string            `mapstructure:"docker_log_path"`
	FileMap       map[string]string // file 与 containerName 之间的映射
}

func NewInputs(inputData map[string]interface{}) (map[string]Input, error) {
	inputs := make(map[string]Input)
	for label, inputObj := range inputData {
		input := Input{}
		input.FileMap = make(map[string]string)
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
		if input.Type == "file" {
			input.Files, err = filepath.Glob(input.Path)
			if err != nil {
				return nil, err
			}
		} else if input.Type == "docker" {
			cts, err := utils.ContainerList(input.CName)
			if err != nil {
				return nil, err
			}
			for _, ct := range cts {
				filePath := strings.Replace(input.DockerLogPath, "{ID}", ct.ID, -1)
				input.Files = append(input.Files, filePath)
				input.FileMap[filePath] = ct.Name
			}
		}
		inputs[label] = input
	}
	return inputs, nil
}
