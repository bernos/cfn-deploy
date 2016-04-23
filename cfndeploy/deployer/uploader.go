package deployer

// Uploader describes an interface for uploading files and folders
// to S3
type Uploader interface {
	UploadFolder(folder string, bucket string, bucketFolder string) UploadResults
	UploadFile(file string, bucket string, key string) UploadResult
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

// UploadResult represents the result of uploading a single file
type UploadResult struct {
	URL   string
	File  string
	Error error
}
