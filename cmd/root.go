// Copyright (c) 2017 off-sync
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logger *logrus.Logger

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "platform-proxy-aws",
	Short: "The Off-Sync.com Platform Proxy for Amazon Web Services",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(
		initConfig,
		initLog)

	// Global flags
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "C", "", "config file (default $HOME/.platform-proxy-aws.yaml)")

	RootCmd.PersistentFlags().StringP("log-level", "L", "Info", "log level")
	viper.SetDefault("logLevel", "Info")
	viper.BindPFlag("logLevel", RootCmd.PersistentFlags().Lookup("log-level"))

	RootCmd.PersistentFlags().BoolP("log-json", "J", true, "use JSON log format")
	viper.SetDefault("logJSON", true)
	viper.BindPFlag("logJSON", RootCmd.PersistentFlags().Lookup("log-json"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".platform-proxy-aws" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".platform-proxy-aws")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initLog() {
	logLevel := viper.GetString("logLevel")

	l, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Printf("Invalid log level '%s': using default log level 'Info'", logLevel)
		l = logrus.InfoLevel
	}

	logger = logrus.New()
	logger.Level = l

	if viper.GetBool("logJSON") {
		logger.Formatter = &logrus.JSONFormatter{}
	}
}
