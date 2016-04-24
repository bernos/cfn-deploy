package uploader

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"os"
	"path"
	"path/filepath"
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
	UploadFolder(folder, bucket, keyPrefix string) (UploadResults, error)
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

func (u *uploader) UploadFolder(folder, bucket, keyPrefix string) (UploadResults, error) {
	var (
		results UploadResults
		wg      sync.WaitGroup
	)

	rc := make(chan *UploadResult)
	defer close(rc)

	files, err := u.buildUploadList(folder)

	if err == nil {
		for _, file := range files {
			wg.Add(1)

			go func(file string) {
				rc <- u.UploadFile(file, bucket, path.Join(keyPrefix, file))
			}(file)
		}

		go func() {
			for r := range rc {
				results = append(results, r)
				wg.Done()
			}
		}()

		wg.Wait()
	}

	if results.HasErrors() {
		err = ErrFolderUpload
	}

	return results, err
}

func (u *uploader) UploadFile(file, bucket, key string) *UploadResult {
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

func (u *uploader) buildUploadList(dir string) ([]string, error) {
	var files []string

	return files, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
}
