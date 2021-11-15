/*
Copyright © 2021 Sourik Ghosh <sourikghosh31@gmail.com>

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
	"apex/internal/upload"
	"apex/pkg/config"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "uplaods all file/s concurrently to scyllaDB",
	Long: `Upload takes --flag dir to find the file/s to upload. For example:

	apex upload --dir "input-path"  it can be relative path as well
	apex upload -d "input-path"
	both will yield the same result.`,
	Run: func(cmd *cobra.Command, args []string) {
		dirName, err := cmd.Flags().GetString("dir")
		if err != nil || dirName == "" {
			fmt.Println("invalid dirname provided")
			return
		}

		logger, err := zap.NewProduction()
		if err != nil {
			fmt.Printf("✋🏾 logger init failed %v", err.Error())
			os.Exit(2)
		}
		defer logger.Sync()

		conn, err := grpc.Dial(config.ServerAddress, grpc.WithInsecure())
		if err != nil {
			logger.Fatal("cannot dial server: ", zap.Error(err))
		}

		if err = uploadFiles(context.Background(), conn, dirName, logger); err != nil {
			logger.Error("", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	// upload cmd Flags
	uploadCmd.Flags().StringP("dir", "d", "~", "Absolute Path to search for file/s to upload")
	uploadCmd.MarkFlagRequired("dir")
}

func uploadFiles(ctx context.Context, cc *grpc.ClientConn, dir string, logger *zap.Logger) error {
	cli := upload.NewUploadClient(ctx, cc, logger, dir)
	var errFailedReqs error

	// reading all the files in the dir
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Fatal("unable to read", zap.String("path", dir), zap.Error(err))
	}
	defer cli.Stop()

	// putting the file to request channel if its not dir
	go func() {
		for _, file := range files {
			if !file.IsDir() {
				cli.Do(filepath.Join(dir, file.Name()))
			}
		}
	}()

	// ranging over each file to check the upload status
	for _, file := range files {
		if !file.IsDir() {
			select {
			case <-cli.DoneRequest:

			case req := <-cli.FailRequest:
				fmt.Println("failed to send " + req)
				errFailedReqs = errors.Wrapf(errFailedReqs, " Failed to send %s", req)
			}
		}
	}

	return errFailedReqs
}
