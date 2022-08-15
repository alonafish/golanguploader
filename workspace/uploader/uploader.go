package uploader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

const (
	AWS_S3_REGION = "eu-west-1"
	AWS_S3_BUCKET = "uploaderbucket123"
)

// file will be uploaded to AWS
func Upload(file multipart.File, fileHeader *multipart.FileHeader, expirationTime int64) (string, error) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(AWS_S3_REGION)})
	if err != nil {
		fmt.Println("Failed to get aws session", err)
		log.Fatal(err)
	}

	// generate new uuid for image name
	uniqueId := uuid.New()
	// remove "- from imageName"
	fileName := strings.Replace(uniqueId.String(), "-", "", -1)

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Println("Failed to get bytes from file")
		return "", err
	}

	_, err = s3.New(session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(AWS_S3_BUCKET),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buf.Bytes()),
		ContentLength:        aws.Int64(int64(buf.Len())),
		ContentType:          aws.String(http.DetectContentType(buf.Bytes())),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	setDeleteTimer(expirationTime, fileName)
	return getPresignedUrl(fileName, expirationTime), err
}

func GetObject(fileName string) (string, error) {
	fmt.Println("Get file ", fileName)

	session, err := session.NewSession(&aws.Config{Region: aws.String(AWS_S3_REGION)})
	if err != nil {
		fmt.Println("Failed to get aws session", err)
		log.Fatal(err)
		return "", err
	}

	resp, err := s3.New(session).GetObject(&s3.GetObjectInput{Bucket: aws.String(AWS_S3_BUCKET), Key: aws.String(fileName)})

	if err != nil {
		fmt.Println("Failed to get file", err)
		log.Fatal(err)
		return "", err
	}

	size := int(*resp.ContentLength)
	fmt.Println("File size ", size)

	return getPresignedUrl(fileName, 1), nil
}

func getPresignedUrl(key string, expirationTime int64) string {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(AWS_S3_REGION)},
	)

	// Create S3 service client
	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(time.Duration(expirationTime) * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}

	return urlStr
}

func setDeleteTimer(expirationTime int64, fileName string) {
	fmt.Println("Set timer")
	time.AfterFunc(time.Minute*time.Duration(expirationTime), func() {
		deleteFile(fileName)
	})
	//defer timer.Stop()
}

func deleteFile(fileName string) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(AWS_S3_REGION)})
	if err != nil {
		fmt.Println("Failed to get aws session", err)
		log.Fatal(err)
	}

	fmt.Println("Delete ", fileName)
	_, err = s3.New(session).DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(AWS_S3_BUCKET), Key: aws.String(fileName)})
	if err != nil {
		fmt.Println("Failed to delete", err)
		log.Fatal(err)
	}
}
