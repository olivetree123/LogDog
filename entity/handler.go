package entity

import (
	"errors"
	"github.com/aarzilli/golua/lua"
	"github.com/mitchellh/mapstructure"
	"sync"
)

type Handler struct {
	ScriptPath string     `mapstructure:"script_path"`
	AddData    *sync.Map  `mapstructure:"add_data"`
	Lua        *lua.State `mapstructure:"-"`
}

func NewHandlers(handlerData map[string]interface{}) (map[string]Handler, error) {
	if handlerData == nil {
		return nil, errors.New("handlerData is nil")
	}
	handlers := make(map[string]Handler)
	for label, handlerObj := range handlerData {
		handler := Handler{}
		err := mapstructure.Decode(handlerObj, &handler)
		if err != nil {
			return nil, err
		}
		if handler.ScriptPath != "" {
			handler.Lua = lua.NewState()
			err = handler.Lua.DoFile(handler.ScriptPath)
			if err != nil {
				return nil, err
			}
			handler.Lua.OpenLibs()
		}
		handlers[label] = handler
	}
	return handlers, nil
}
