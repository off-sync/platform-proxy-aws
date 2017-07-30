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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/off-sync/platform-proxy-app/infra/logging"
	"github.com/off-sync/platform-proxy-app/proxies/cmd/startproxy"
	"github.com/off-sync/platform-proxy-aws/infra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the Off-Sync.com Platform Proxy for AWS",
	Long:  ``,
	Run:   run,
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(cmd *cobra.Command, args []string) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(viper.GetString("awsRegion"))})
	if err != nil {
		logger.WithError(err).Fatal("creating new session")
	}

	ecsSvc := ecs.New(sess)

	serviceRepository, err := infra.NewServiceRepository(ecsSvc, viper.GetString("ecsClusterName"))
	if err != nil {
		logger.
			WithError(err).
			Fatal("creating service repository")

		return
	}

	startProxyCmd, err := startproxy.NewCommand(
		serviceRepository,
		nil,
		logging.NewLogrusLogger(logger))
	if err != nil {
		logger.WithError(err).Fatal("creating start proxy command")

		return
	}

	startProxyCmd.Execute(&startproxy.Model{})
}
