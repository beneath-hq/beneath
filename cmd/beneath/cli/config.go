package cli

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/spf13/cobra"
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

		// eg. if BENEATH_ENV=dev, then loads config/.development.yaml
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

		// bind to env variable name
		envVarName := c.configKeyToEnvVarName(configKey.Key)
		c.v.BindEnv(configKey.Key, envVarName)
	}

	// necessary when using UnmarshalKey and loading env keys (see https://github.com/spf13/viper/issues/798)
	for _, key := range c.v.AllKeys() {
		c.v.Set(key, c.v.Get(key))
	}

	return nil
}

// converts a config key to an env variable name (eg. "mq_driver" -> "BENEATH_MQ_DRIVER")
func (c *CLI) configKeyToEnvVarName(key string) string {
	envVarSuffix := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	envVarName := fmt.Sprintf("%s_%s", envVarPrefix, envVarSuffix)
	return envVarName
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

// registers command that generates a config template based on registered ConfigKeys
func (c *CLI) newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:    "config",
		Short:  "Helpers for config files (internal: for use during development)",
		Hidden: envutil.GetEnv() == envutil.Production,
	}
	configCmd.AddCommand(&cobra.Command{
		Use:   "generate FILE",
		Short: "(Re)generates a default config.yaml template based on registered config keys",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c.generateConfigTemplate(args[0])
		},
	})
	return configCmd
}

func (c *CLI) generateConfigTemplate(file string) {
	tree, err := c.configKeysToTree(configKeys)
	if err != nil {
		panic(err)
	}
	yaml := c.configKeyNodeToYaml(tree, 0)
	err = ioutil.WriteFile(file, []byte(yaml), 0644)
	if err != nil {
		panic(err)
	}
}

func (c *CLI) configKeysToTree(configKeys []*ConfigKey) (map[string]interface{}, error) {
	tree := make(map[string]interface{})
	for _, key := range configKeys {
		node := tree
		path := strings.Split(key.Key, ".")
		for j, component := range path {
			last := j+1 == len(path)
			if last {
				node[component] = key
			} else {
				sub, ok := node[component]
				if !ok {
					subMap := make(map[string]interface{})
					node[component] = subMap
					node = subMap
				} else {
					subMap, ok := sub.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("ambiguous key '%s' is both a value and parent", strings.Join(path[:j], "."))
					}
					node = subMap
				}
			}
		}
	}
	return tree, nil
}

func (c *CLI) configKeyNodeToYaml(node interface{}, indent int) string {
	pad := strings.Repeat(" ", indent)
	switch node := node.(type) {
	case *ConfigKey:
		res := pad + "# "
		if node.Description != "" {
			res += node.Description + " "
		}
		res += fmt.Sprintf("(env: %s)", c.configKeyToEnvVarName(node.Key))
		res += "\n"
		components := strings.Split(node.Key, ".")
		res += pad + components[len(components)-1] + ": " + `"` + fmt.Sprintf("%v", node.Default) + `"` + "\n"
		return res
	case map[string]interface{}:
		keys := make([]string, 0, len(node))
		for key := range node {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		res := ""
		for _, key := range keys {
			sub := node[key]
			if _, ok := sub.(*ConfigKey); ok {
				res += c.configKeyNodeToYaml(sub, indent)
			} else {
				res += pad + key + ":\n" + c.configKeyNodeToYaml(sub, indent+2)
			}
			if indent == 0 {
				res += "\n"
			}
		}
		return res
	default:
		panic(fmt.Errorf("unexpected node type '%T'", node))
	}
}
