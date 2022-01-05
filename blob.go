package passport

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/h2non/filetype"
	"github.com/ninja-software/terror/v2"
)

// Blob is a single attachment item on the platform
type Blob struct {
	ID            BlobID     `json:"id" db:"id"`
	FileName      string     `json:"fileName" db:"file_name"`
	MimeType      string     `json:"mimeType" db:"mime_type"`
	FileSizeBytes int64      `json:"fileSizeBytes" db:"file_size_bytes"`
	Extension     string     `json:"extension" db:"extension"`
	File          []byte     `json:"file" db:"file"`
	Views         int        `json:"views" db:"views"`
	Hash          *string    `json:"hash" db:"hash"`
	Public        bool       `json:"public" db:"public"`
	DeletedAt     *time.Time `json:"deletedAt" db:"deleted_at"`
	UpdateAt      *time.Time `json:"updatedAt" db:"updated_at"`
	CreatedAt     *time.Time `json:"createdAt" db:"created_at"`
}

// BlobFromURL downloads an image from a url and returns a blob
func BlobFromURL(url string, fileName string) (*Blob, error) {
	// Get Image
	resp, err := http.Get(url)
	if err != nil {
		return nil, terror.Error(err, "get file from url")
	}
	defer resp.Body.Close()

	// Read image
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, terror.Error(terror.ErrParse, "parse error")
	}

	// Get image mime type
	kind, err := filetype.Match(data)
	if err != nil {
		return nil, terror.Error(terror.ErrParse, "parse error")
	}

	if kind == filetype.Unknown {
		return nil, terror.Error(fmt.Errorf("Image type is unknown"), "Image type is unknown")
	}

	mimeType := kind.MIME.Value
	extension := kind.Extension

	// Get hash
	hasher := md5.New()
	_, err = hasher.Write(data)
	if err != nil {
		return nil, terror.Error(err, "hash error")
	}
	hashResult := hasher.Sum(nil)
	hash := hex.EncodeToString(hashResult)

	// Create image blob
	image := &Blob{
		FileName:      fileName,
		MimeType:      mimeType,
		Extension:     extension,
		FileSizeBytes: int64(len(data)),
		File:          data,
		Hash:          &hash,
	}

	return image, nil
}
