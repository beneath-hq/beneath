package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
)

// prefix for environment variables that override config keys
const envVarPrefix = "BENEATH"

// loads config based on the given file, or if it's not set, the current environment
func (c *CLI) loadConfig(configFileOrEmpty string) error {
	// load config file
	if configFileOrEmpty != "" {
		err := c.loadConfigFile(configFileOrEmpty)
		if err != nil {
			return err
		}
	} else {
		// always loads config/.default
		err := c.searchLoadAndMergeConfigName(".default")
		if err != nil {
			return err
		}

		// eg. if ENV=dev, then loads config/.development.yaml
		err = c.searchLoadAndMergeConfigName(fmt.Sprintf(".%s", envutil.GetEnv()))
		if err != nil {
			return err
		}
	}

	// Not using v.AutomaticEnv() and v.SetEnvPrefix() to directly handle
	// conversion of key names to env variable names.

	// register every config key
	for _, configKey := range configKeys { // configKeys set in registry.go
		// set default
		c.v.SetDefault(configKey.Key, configKey.Default)

		// bind to env variable name (eg. "mq_driver" -> "BENEATH_MQ_DRIVER")
		envVarSuffix := strings.ToUpper(strings.ReplaceAll(configKey.Key, ".", "_"))
		envVarName := fmt.Sprintf("%s_%s", envVarPrefix, envVarSuffix)
		c.v.BindEnv(configKey.Key, envVarName)
	}

	return nil
}

// loads config file (errors if file doesn't exist)
func (c *CLI) loadConfigFile(file string) error {
	c.v.SetConfigFile(file)
	if err := c.v.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

// searches for a config file with the given name, and merges it
// into c.v if found
func (c *CLI) searchLoadAndMergeConfigName(name string) error {
	// create new viper instance
	v := viper.New()

	// configure to look for file with name
	v.SetConfigName(name)
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// search for config file (don't error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		return nil
	}

	// merge into c.v (main viper object)
	c.v.MergeConfigMap(v.AllSettings())
	return nil
}
