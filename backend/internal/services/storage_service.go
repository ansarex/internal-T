package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/trustwired/internal-t/internal/config"
)

type StorageService struct {
	Config *appconfig.Config
	s3     *s3.Client
}

func NewStorageService(cfg *appconfig.Config) (*StorageService, error) {
	svc := &StorageService{Config: cfg}

	if cfg.FilesystemDisk == "s3" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           cfg.AWSEndpoint,
				SigningRegion: cfg.AWSRegion,
			}, nil
		})

		awsCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.AWSRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKeyID,
				cfg.AWSSecretKey,
				"",
			)),
			config.WithEndpointResolverWithOptions(customResolver),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to configure S3: %w", err)
		}

		svc.s3 = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	} else {
		if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	return svc, nil
}

func (s *StorageService) Store(file multipart.File, storagePath string) error {
	if s.Config.FilesystemDisk == "s3" {
		return s.storeS3(file, storagePath)
	}
	return s.storeLocal(file, storagePath)
}

func (s *StorageService) storeLocal(file multipart.File, storagePath string) error {
	fullPath := filepath.Join(s.Config.StoragePath, storagePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	dst, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	return err
}

func (s *StorageService) storeS3(file multipart.File, storagePath string) error {
	_, err := s.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.Config.AWSBucket),
		Key:    aws.String(storagePath),
		Body:   file,
	})
	return err
}

func (s *StorageService) Get(storagePath string) (io.ReadCloser, error) {
	if s.Config.FilesystemDisk == "s3" {
		return s.getS3(storagePath)
	}
	return s.getLocal(storagePath)
}

func (s *StorageService) getLocal(storagePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.Config.StoragePath, storagePath)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return f, nil
}

func (s *StorageService) getS3(storagePath string) (io.ReadCloser, error) {
	result, err := s.s3.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.Config.AWSBucket),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	return result.Body, nil
}

func (s *StorageService) Delete(storagePath string) error {
	if s.Config.FilesystemDisk == "s3" {
		_, err := s.s3.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(s.Config.AWSBucket),
			Key:    aws.String(storagePath),
		})
		return err
	}
	return os.Remove(filepath.Join(s.Config.StoragePath, storagePath))
}

func GenerateAgreementPath(jobID uint, filename string) string {
	ext := filepath.Ext(filename)
	safe := sanitizeFilename(strings.TrimSuffix(filename, ext))
	ts := time.Now().Unix()
	return fmt.Sprintf("agreements/job-%d/%s_%d%s", jobID, safe, ts, ext)
}

func GenerateSignedCopyPath(jobID uint, filename string) string {
	ext := filepath.Ext(filename)
	ts := time.Now().Unix()
	return fmt.Sprintf("signed-copies/job-%d/signed_%d%s", jobID, ts, ext)
}

func GenerateReceiptPath(billingMonth time.Time, filename string) string {
	ext := filepath.Ext(filename)
	safe := sanitizeFilename(strings.TrimSuffix(filename, ext))
	ts := time.Now().Unix()
	month := billingMonth.Format("2006-01-02")
	return fmt.Sprintf("receipts/%s/%s_%d%s", month, safe, ts, ext)
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", "..", "_")
	return replacer.Replace(name)
}
