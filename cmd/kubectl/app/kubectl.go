package app

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	// Used for flags.
	cfgFile string
	baseUrl string

	rootCmd = &cobra.Command{
		Use:   "odin",
		Short: "A kubectl for minik8s",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "podConfig", "", "podConfig file (default is $HOME/.odin.yaml)")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".odin")
		viper.SetConfigType("yaml")
	}
	viper.SetDefault("url", "http://localhost:8080")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using podConfig file: %s\n", viper.ConfigFileUsed())
	}
	baseUrl = viper.GetString("url")
}
