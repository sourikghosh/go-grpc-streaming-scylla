/*
Copyright © 2021 Sourik Ghosh sourikghosh31@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"apex/pkg/config"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apex",
	Short: "Apex allows to multi-upload file/s concurrently to scyllaDB",
	Long: `Apex searchs all file/s in the input directory to concurrently upload them to scyllaDB.
It was a project to get familiarize with gRPC streams and scyllaDB.Apex takes two config --flag. For example:

If the gRPC server is not running on localhost:8080 you can change it with
	apex --addr localhost:8080

You can change the concurrency with 
	apex --workerCount 6
Higher workerCount means higher concurrency
The default value is 6.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// rootFlags
	rootCmd.PersistentFlags().StringVarP(&config.ServerAddress, "addr", "a", "localhost:1500", "the server address")
	rootCmd.PersistentFlags().IntVarP(&config.MaxWorkerCount, "workerCount", "w", 6, "no of concurrent worker count to upload files")
}

func initConfig() {

}
