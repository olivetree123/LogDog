package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io"
	. "logDog/common"
	"logDog/entity"
	"logDog/process"
	"os"
	"regexp"
	"strings"
	"sync"
)

var MainConfig *viper.Viper
var MetaConfig *viper.Viper
var Lock sync.Mutex
var Inputs map[string]entity.Input
var Outputs map[string]entity.Output
var Handlers map[string]entity.Handler
var cmdHelp, cmdVersion bool
var cmdConfig string
var MainConfigDirPath string
var MetaConfigDirPath string
var MainConfigFile string
var MetaConfigFile string

func init() {
	MainConfigDirPath = "/etc/logdog/"
	MetaConfigDirPath = "/var/lib/logdog/"
	// MainConfigFile 默认值为 /etc/logdog/logdog.toml。但是你可以通过命令行自定义指定。
	MainConfigFile = MainConfigDirPath + "logdog.toml"
	// MetaConfigDirPath 默认值为 /var/lib/logdog/logdog_meta.toml，暂时不可以自定义。
	MetaConfigFile = MetaConfigDirPath + "logdog_meta.toml"
	flag.BoolVar(&cmdHelp, "h", false, "this help")
	flag.BoolVar(&cmdVersion, "v", false, "show version")
	flag.StringVar(&cmdConfig, "c", "/etc/logdog/logdog.toml", "set config file")
}

func prepareMetaConfigFile() error {
	if _, err := os.Stat(MetaConfigDirPath); os.IsNotExist(err) {
		err := os.Mkdir(MetaConfigDirPath, 0777)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(MetaConfigFile); os.IsNotExist(err) {
		f, err := os.Create(MetaConfigFile)
		if err != nil {
			return err
		}
		_, err = f.WriteString("[rows]")
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func loadMetaConfig() error {
	MetaConfig = viper.New()
	MetaConfig.SetConfigFile(MetaConfigFile)
	err := MetaConfig.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

func loadMainConfig() error {
	MainConfig = viper.New()
	MainConfig.SetConfigFile(MainConfigFile)
	err := MainConfig.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

func getMetaRowKey(filePath string) string {
	filePath = strings.Replace(filePath, "/", "_", -1)
	filePath = strings.Replace(filePath, ".", "_", -1)
	return fmt.Sprintf("rows.%s", filePath)
}

func getFileRowRead(filePath string) int {
	row := MetaConfig.Get(getMetaRowKey(filePath))
	if row == nil {
		return 0
	}
	return int(row.(int64))
}

func writeMetaData(filePath string, lineNum int) error {
	MetaConfig.Set(getMetaRowKey(filePath), lineNum)
	err := MetaConfig.WriteConfig()
	if err != nil {
		Logger.Error(err)
		return err
	}
	return nil
}

func readFile(label string) {
	input := Inputs[label]
	var wg sync.WaitGroup
	for _, filePath := range input.Files {
		wg.Add(1)
		go func(filePath string) {
			// 从第一行开始读
			row := 1
			fileWatcher := entity.NewFileWatcher(filePath)
			reader := bufio.NewReader(fileWatcher.File)
			signalChannel := make(chan bool)
			rowRead := getFileRowRead(filePath)
			for {
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							break
						}
						Logger.Error(err)
						break
					}
					line = strings.Trim(line, "\n")
					if line == "" {
						continue
					}
					if row <= rowRead {
						row++
						Logger.Info("continue")
						continue
					}
					//var containerName string
					//if input.FileMap != nil {
					//	containerName = input.FileMap[filePath]
					//}
					message, err := parseMessage(filePath, line, input.Format, input.CompiledRegex, input.FileMap[filePath])
					if err != nil {
						Logger.Error(err)
						signalChannel <- true
						return
					}
					message = handler(label, message)
					err = output(label, message)
					if err != nil {
						Logger.Error(err)
						signalChannel <- true
						return
					}
					err = writeMetaData(filePath, row)
					if err != nil {
						Logger.Error(err)
						signalChannel <- true
						return
					}

					row++
				}
				select {
				case event := <-fileWatcher.Watcher.Events:
					{
						if event.Op&fsnotify.Create == fsnotify.Create {
							// 1. 创建文件
							Logger.Info("Create File.")
						} else if event.Op&fsnotify.Write == fsnotify.Write {
							// 2. 写文件
							Logger.Info("Write File.")
							fi, err := fileWatcher.File.Stat()
							if err != nil {
								if os.IsNotExist(err) {
									err = fileWatcher.ReWatch()
									if err != nil {
										Logger.Error(err)
										return
									}
								} else {
									Logger.Error(err)
									return
								}
							}
							// file was truncated
							if fileWatcher.Offset > fi.Size() {
								err = fileWatcher.ReWatch()
								if err != nil {
									Logger.Error(err)
									return
								}
							}
							continue
						} else if event.Op&fsnotify.Remove == fsnotify.Remove {
							// 3. 删除文件
							Logger.Info("Remove File.")
						} else if event.Op&fsnotify.Rename == fsnotify.Rename {
							// 4. 重命名文件
							Logger.Info("Rename File.")
						} else if event.Op&fsnotify.Chmod == fsnotify.Chmod {
							// 5. 修改文件权限
							Logger.Info("Chmod File.")
						}
					}
				case err, ok := <-fileWatcher.Watcher.Errors:
					{
						if !ok {
							Logger.Error(ok)
							return
						}
						Logger.Error(err)
					}
				case <-signalChannel:
					{
						wg.Done()
						return
					}
				}
			}
		}(filePath)
	}
	wg.Wait()
}

func handler(label string, message map[string]interface{}) map[string]interface{} {
	handler := Handlers[label]
	if handler.AddData != nil {
		r := make(map[string]interface{})
		handler.AddData.Range(func(key, val interface{}) bool {
			r[key.(string)] = val
			return true
		})
		message["add_data"] = r
	}
	if handler.ScriptPath != "" {
		message = runLuaScript(handler.Lua, message)
	}
	return message
}

func parseMessage(filePath string, message string, format string, regex *regexp.Regexp, containerName string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["_source"] = filePath
	result["_container"] = containerName
	if format == "json" {
		var data interface{}
		err := json.Unmarshal([]byte(message), &data)
		if err != nil {
			return nil, err
		}
		result["message"] = data
	} else if format == "text" {
		if regex == nil {
			result["message"] = message
		} else {
			match := regex.FindStringSubmatch(message)
			r := make(map[string]string)
			for i, name := range match {
				if i == 0 {
					continue
				}
				key := regex.SubexpNames()[i]
				if key == "" {
					return nil, errors.New("invalid regex")
				}
				r[key] = name
			}
			result["message"] = r
		}
	} else {
		msg := fmt.Sprintf("invalid format = %s", format)
		return nil, errors.New(msg)
	}
	return result, nil
}

func output(label string, message map[string]interface{}) error {
	var err error
	output := Outputs[label]
	message["_hostname"] = output.HostName
	if output.Type == "stdout" {
		Logger.Info(message)
	} else if output.Type == "redis" {
		err = process.OutputRedis(output, label, message)
	} else if output.Type == "http" {
		err = process.OutputHttp(output, message)
	}
	return err
}

func runLuaScript(Lua *lua.State, data map[string]interface{}) map[string]interface{} {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		Logger.Error(err)
		return nil
	}
	Lock.Lock()
	Lua.GetField(lua.LUA_GLOBALSINDEX, "myHandler")
	Lua.PushBytes(dataBytes)
	err = Lua.Call(1, 1)
	if err != nil {
		Logger.Error(err)
		return nil
	}
	ret := Lua.ToBytes(1)
	r := make(map[string]interface{})
	err = json.Unmarshal(ret, &r)
	if err != nil {
		Logger.Error(err)
		return nil
	}
	Lua.Pop(1)
	Lock.Unlock()
	return r
}

func main() {
	flag.Parse()
	if cmdHelp {
		Logger.Info("Help")
		flag.Usage()
		return
	}
	if cmdVersion {
		Logger.Info("Version")
		return
	}
	if cmdConfig != "" {
		if _, err := os.Stat(cmdConfig); os.IsNotExist(err) {
			Logger.Error(err)
			return
		}
		MainConfigFile = cmdConfig
	}
	err := prepareMetaConfigFile()
	if err != nil {
		Logger.Error(err)
		return
	}
	err = loadMainConfig()
	if err != nil {
		Logger.Error(err)
		return
	}
	err = loadMetaConfig()
	if err != nil {
		Logger.Error(err)
		return
	}
	inputData := MainConfig.Get("input")
	outputData := MainConfig.Get("output")
	handlerData := MainConfig.Get("handler")
	if inputData == nil {
		Logger.Error("no input found")
		return
	}
	if outputData == nil {
		Logger.Error("no output found")
		return
	}
	Inputs, err = entity.NewInputs(inputData.(map[string]interface{}))
	if err != nil {
		Logger.Error(err)
		return
	}
	Outputs, err = entity.NewOutputs(outputData.(map[string]interface{}))
	if err != nil {
		Logger.Error(err)
		return
	}
	Handlers, err = entity.NewHandlers(handlerData.(map[string]interface{}))
	if err != nil {
		Logger.Error(err)
		return
	}
	for label := range Inputs {
		readFile(label)
	}
}
