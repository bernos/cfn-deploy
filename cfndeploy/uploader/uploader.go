package uploader

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	// ErrFolderUpload means that one or more file uploads failed when uploading
	// a folder.
	ErrFolderUpload = errors.New("text string")
)

// Uploader describes an interface for uploading files and folders
// to S3
type Uploader interface {
	UploadFiles(files []string, basePath, bucket, keyPrefix string) (UploadResults, error)
	UploadFile(file, bucket, key string) *UploadResult
}

// UploadResults represents the result of uploading multiple files
type UploadResults []*UploadResult

// HasErrors returns true if any upload results contained errors
func (results UploadResults) HasErrors() bool {
	for i := range results {
		if results[i].Error != nil {
			return true
		}
	}
	return false
}

// Errors returns a slice of all errors encountered when uploading
// a folder
func (results UploadResults) Errors() []error {
	var errors []error
	for i := range results {
		if results[i].Error != nil {
			errors = append(errors, results[i].Error)
		}
	}
	return errors
}

// UploadResult represents the result of uploading a single file
type UploadResult struct {
	URL   string
	File  string
	Error error
}

// New builds a new S3 uploader
func New(s3 s3manageriface.UploaderAPI) Uploader {
	return &uploader{
		s3: s3,
	}
}

type uploader struct {
	s3 s3manageriface.UploaderAPI
}

func (u *uploader) UploadFiles(files []string, basePath, bucket, keyPrefix string) (UploadResults, error) {
	var (
		results UploadResults
		wg      sync.WaitGroup
	)

	rc := make(chan *UploadResult)
	defer close(rc)

	for _, file := range files {
		wg.Add(1)

		go func(file string) {
			key, err := calculateBucketKey(file, basePath, keyPrefix)

			if err != nil {
				log.Printf("Error calculating bucket key: %s", err.Error())
				rc <- &UploadResult{Error: err}
			} else {
				rc <- u.UploadFile(file, bucket, key)
			}
		}(file)
	}

	go func() {
		for r := range rc {
			results = append(results, r)
			wg.Done()
		}
	}()

	wg.Wait()

	if results.HasErrors() {
		var msg string
		for _, err := range results.Errors() {
			msg = msg + ", " + err.Error()
		}
		return results, fmt.Errorf("%s", msg)
	}

	return results, nil
}

func calculateBucketKey(file, basePath, prefix string) (string, error) {
	file, err := stripBasePath(file, basePath)

	if err != nil {
		return file, err
	}

	return path.Join(prefix, file), nil
}

func stripBasePath(file, basePath string) (string, error) {
	if strings.Index(file, basePath) == 0 {
		return file[len(basePath):], nil
	}
	return file, fmt.Errorf("File %s not based at %s", file, basePath)
}

func (u *uploader) UploadFile(file, bucket, key string) *UploadResult {
	log.Printf("UploadFile(%s, %s, %s)", file, bucket, key)
	result := &UploadResult{
		File: file,
	}

	f, err := os.Open(file)

	if err != nil {
		result.Error = err
		return result
	}

	defer f.Close()

	options := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	}

	resp, err := u.s3.Upload(options)

	if err != nil {
		result.Error = err
	} else {
		result.URL = resp.Location
	}

	return result
}
