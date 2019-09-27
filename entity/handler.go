package entity

import (
	"github.com/aarzilli/golua/lua"
	"github.com/mitchellh/mapstructure"
	. "logDog/common"
	"sync"
)

type Handler struct {
	ScriptPath string     `mapstructure:"script_path"`
	AddData    *sync.Map  `mapstructure:"add_data"`
	Lua        *lua.State `mapstructure:"-"`
}

func NewHandlers(handlerData map[string]interface{}) map[string]Handler {
	handlers := make(map[string]Handler)
	for label, handlerObj := range handlerData {
		handler := Handler{}
		err := mapstructure.Decode(handlerObj, &handler)
		if err != nil {
			Logger.Error(err)
			return nil
		}
		if handler.ScriptPath != "" {
			handler.Lua = lua.NewState()
			err = handler.Lua.DoFile(handler.ScriptPath)
			if err != nil {
				Logger.Error(err)
				return nil
			}
			handler.Lua.OpenLibs()
		}
		handlers[label] = handler
	}
	return handlers
}
