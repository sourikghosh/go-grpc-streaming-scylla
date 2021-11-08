package config

import "go.uber.org/zap"

const (
	MiB100            float64 = 1e8
	MiB5              float64 = 5e6
	MaxUploadFileSize float64 = MiB5
)

var ZapLogger *zap.Logger
