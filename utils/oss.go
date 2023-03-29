package utils

import (
	"bytes"
	"fmt"
	"ggslayer/utils/config"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"os"
)

type OssClient struct {
	Client *oss.Client
}

func NewClient() *OssClient {
	endpoint := config.GetString("oss.Url")
	accessID := config.GetString("oss.AccessID")
	accessKey := config.GetString("oss.AccessKey")
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return &OssClient{Client: client}
}

//GetBucket 创建bucket
func (s *OssClient) GetBucket(bucketName string) (bucket *oss.Bucket, err error) {
	err = s.Client.CreateBucket(bucketName)
	if err != nil {
		logrus.Error(err)
		return
	}
	// Get bucket
	bucket, err = s.Client.Bucket(bucketName)
	if err != nil {
		logrus.Error(err)
	}
	return
}

//UploadImage 上传图片
func (s *OssClient) UploadImage(objectName string, fileByte []byte) error {
	bucket, err := s.Client.Bucket(config.GetString("oss.Bucket")) //注意此处不要写错，写错的话，err让然是nil，我们应该需要先判断一下是否存在
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = bucket.PutObject(objectName, bytes.NewReader(fileByte))
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (s *OssClient) ListBucket() {
	marker := ""
	for {
		lsRes, err := s.Client.ListBuckets(oss.Marker(marker))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}

		// 默认情况下一次返回100条记录。
		for _, bucket := range lsRes.Buckets {
			fmt.Println("Bucket: ", bucket.Name)
		}
		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}
}
