package pkg

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"airbox/config"
)

var oss *minio.Client

func GetOSS() *minio.Client {
	return oss
}

// InitializeOSS 初始化对象存储
func InitializeOSS() error {
	// Initialize minio client object.
	var err error
	oss, err = minio.New(config.GetConfig().MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.GetConfig().MinIO.AccessKey, config.GetConfig().MinIO.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}
	return nil
}
