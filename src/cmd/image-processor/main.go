package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
	"io"
	"log"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
)

type SFNEvent struct {
	BatchInput BatchInput `json:"BatchInput"`
	Items      []Item     `json:"Items"`
}
type BatchInput struct {
	LambdaProcessorArn string `json:"lambda_processor_arn"`
	SourceBucketName   string `json:"source_bucket_name"`
	DestBucketName     string `json:"destination_bucket_name"`
}
type Item struct {
	Key  string `json:"Key"`
	Etag string `json:"Etag"`
}

func HandleRequest(ctx context.Context, customEvent SFNEvent) {
	body, err := json.MarshalIndent(customEvent, " ", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(body))
	log.Printf("Running image-processor on source bucket: %s, lambda_arn: %s", customEvent.BatchInput.SourceBucketName, customEvent.BatchInput.LambdaProcessorArn)
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal("Cannot create s3 client")
	}
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	wg := &sync.WaitGroup{}
	errorsChan := make(chan bool)

	go func(errorsChan chan bool) {
		count := 0
		for <-errorsChan {
			count++
			if count > len(customEvent.Items)/2 {
				log.Fatal("Too many errors")
			}
		}
	}(errorsChan)

	for _, item := range customEvent.Items {
		log.Printf("Processing item: %s", item.Key)
		wg.Add(1)
		go ProcessImage(wg, item.Key, customEvent.BatchInput.SourceBucketName, customEvent.BatchInput.DestBucketName, uploader, downloader, errorsChan)
	}
	wg.Wait()
	log.Println("Processing is complete")
	return
}

func ProcessImage(wg *sync.WaitGroup, key, sourceBucket, destBucket string, uploader *s3manager.Uploader, downloader *s3manager.Downloader, errorsChan chan bool) {
	srcImage, err := DownloadImage(downloader, sourceBucket, key)
	if err != nil {
		log.Printf("Failed image %s during download, %s", key, err)
		errorsChan <- true
		wg.Done()
		return
	}
	dstImage, err := ResizeImage(srcImage)
	if err != nil {
		log.Printf("Failed image %s during resizing, %s", key, err)
		errorsChan <- true
		wg.Done()
		return
	}
	uploadID, err := UploadImage(uploader, destBucket, key, dstImage)
	if err != nil {
		log.Printf("Failed image %s during uploading, %s", key, err)
		errorsChan <- true
		wg.Done()
		return
	}
	log.Printf("Upload completed successfully for key %s with uploadID %s", key, uploadID)
	wg.Done()
}

func DownloadImage(downloader *s3manager.Downloader, bucket, key string) (io.Reader, error) {
	var buff aws.WriteAtBuffer
	_, err := downloader.Download(&buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buff.Bytes()), nil
}

func ResizeImage(data io.Reader) (io.Reader, error) {
	sourceImage, _, err := image.Decode(data)
	if err != nil {
		return nil, err
	}
	dstImage128 := imaging.Resize(sourceImage, 128, 128, imaging.Lanczos)
	var destBuffer bytes.Buffer
	destWriter := bufio.NewWriter(&destBuffer)
	err = jpeg.Encode(destWriter, dstImage128, nil)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(destBuffer.Bytes())
	return reader, nil
}

func UploadImage(uploader *s3manager.Uploader, bucket, key string, data io.Reader) (string, error) {
	res, err := uploader.Upload(&s3manager.UploadInput{
		Key:    aws.String(key),
		Bucket: aws.String(bucket),
		Body:   data,
	})
	if err != nil {
		return "", err
	}
	return res.UploadID, nil
}

func main() {
	lambda.Start(HandleRequest)
}
