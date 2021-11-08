package config

import "go.uber.org/zap"

const (
	MiB100            float64 = 1e8
	MiB5              float64 = 5e6
	MiB2              float64 = 2e6
	MaxUploadFileSize float64 = MiB2
)

var ZapLogger *zap.Logger
