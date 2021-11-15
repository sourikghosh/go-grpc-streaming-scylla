package config

const (
	MiB100            float64 = 1e8
	MiB10             float64 = 1e7
	MiB5              float64 = 5e6
	MiB1              float64 = 1e6
	MaxUploadFileSize float64 = MiB10
)

var (
	MaxWorkerCount int
	ServerAddress  string
)
