package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml"
)

const sysEnvKeyAppConfig = "APP_CONFIG_PATH"

var defaultFileName = "config.toml"

var configFilePath string
var conf *toml.Tree

func init() {
	// Load .env file first
	loadEnvFile()
	LoadLocalConfig(defaultFileName)
}

func loadEnvFile() {
	envPath := filepath.Join(currentPath(), ".env")
	if isFileExist(envPath) {
		log.Println("loading .env file from:", envPath)
		err := godotenv.Load(envPath)
		if err != nil {
			log.Printf("warning: could not load .env file: %v\n", err)
		} else {
			log.Println(".env file loaded successfully")
		}
	} else {
		log.Println(".env file not found, checking exec path")
		execEnvPath := filepath.Join(execPath(), ".env")
		if isFileExist(execEnvPath) {
			log.Println("loading .env file from:", execEnvPath)
			err := godotenv.Load(execEnvPath)
			if err != nil {
				log.Printf("warning: could not load .env file: %v\n", err)
			} else {
				log.Println(".env file loaded successfully")
			}
		}
	}
}
func LoadLocalConfig(configFileName string) {
	log.Println("loading config file")
	defaultFileName = configFileName

	configFilePath = configPath()
	LoadConfigByPath(configFilePath)
}

func LoadConfigByPath(fullPath string) {
	log.Printf("load config by path: %s\n", fullPath)
	if !isFileExist(fullPath) {
		log.Println("config file not found:", fullPath)
		return
	}

	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		fmt.Print(err)
	}
	str := string(b)
	log.Printf("config file content: \n%v", str)

	conf, _ = toml.Load(str)
}

func configPath() string {
	path := os.Getenv(sysEnvKeyAppConfig)
	if path != "" && isFileExist(getConfigFilePath(path)) {
		return configFileFullPath(path)
	}
	path = execPath()
	if path != "" && isFileExist(getConfigFilePath(path)) {
		return configFileFullPath(path)
	}

	path = currentPath()

	fullPath := filepath.Join(path, defaultFileName)
	log.Println("config file path:", fullPath)
	return fullPath
}
func configFileFullPath(path string) string {
	return filepath.Join(path, defaultFileName)
}

func getConfigFilePath(configPath string) string {
	return filepath.Join(configPath, defaultFileName)
}

func GetInt(key string) int64 {
	var foo int64
	foo = -1
	return GetIntWithDefaultValue(key, foo)
}

func GetBool(key string) bool {
	return GetBoolWithDefaultValue(key, false)
}

//func GetStringList(key string) []string {
//	return conf.Get(key).(string)
//}

func GetString(key string, def string) string {
	// First check environment variable
	envKey := toEnvKey(key)
	envValue := os.Getenv(envKey)
	if envValue != "" {
		log.Printf("key: %s, value from env: %s", key, envValue)
		return envValue
	}

	// Fall back to TOML config
	var value string
	if conf == nil {
		value = def
	} else {
		obj := conf.Get(key)
		if obj == nil {
			value = def
		} else {
			value = obj.(string)
		}
	}

	// log.Printf("key: %s, value: %s", key, value)

	return value
}

// toEnvKey converts a dotted key like "redis.address" to "REDIS_ADDRESS"
func toEnvKey(key string) string {
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	return envKey
}

func GetIntWithDefaultValue(key string, def int64) int64 {
	// First check environment variable
	envKey := toEnvKey(key)
	envValue := os.Getenv(envKey)
	if envValue != "" {
		if intVal, err := strconv.ParseInt(envValue, 10, 64); err == nil {
			log.Printf("key: %s, value from env: %v", key, intVal)
			return intVal
		}
	}

	// Fall back to TOML config
	var value int64
	if conf == nil {
		return def
	}

	k := conf.Get(key)
	if k == nil {
		value = def
	} else {
		value = k.(int64)
	}

	log.Printf("key: %s, value: %v", key, value)

	return value
}

func GetBoolWithDefaultValue(key string, def bool) bool {
	// First check environment variable
	envKey := toEnvKey(key)
	envValue := os.Getenv(envKey)
	if envValue != "" {
		if boolVal, err := strconv.ParseBool(envValue); err == nil {
			log.Printf("key: %s, value from env: %v", key, boolVal)
			return boolVal
		}
	}

	// Fall back to TOML config
	var result bool
	obj := conf.Get(key)
	if obj == nil {
		result = def
	} else {
		result = obj.(bool)
	}
	return result
}

func isFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	fileExist := err == nil || os.IsExist(err)

	log.Printf("file: %s, exist:%v", filePath, fileExist)
	return fileExist
}

func execPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func currentPath() string {
	currentPath, _ := os.Getwd()
	return currentPath
}
