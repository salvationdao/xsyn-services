package api

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"xsyn-services/boiler"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/volatiletech/null/v8"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/h2non/filetype"
	"github.com/ninja-software/terror/v2"
)

var (
	ErrUnknownFileType = fmt.Errorf("file type is unknown")
	ErrInvalidFileType = fmt.Errorf("file type is invalid")
)

// FilesController holds connection data for handlers
type FilesController struct {
	API *API
}

// FileRouter returns a new router for handling File requests
func FileRouter(api *API) chi.Router {
	c := &FilesController{
		API: api,
	}

	r := chi.NewRouter()
	r.Get("/{id}", WithError(c.FileGet))
	r.Get("/item/{name}", WithError(c.FileGetByName))
	r.Post("/upload", WithError(WithUser(api, c.FileUpload)))

	return r
}

// FileGet retrives a file attachment
func (c *FilesController) FileGet(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()

	// Get blob id
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "no id provided")
	}
	id, err := uuid.FromString(idStr)
	blobID := types.BlobID(id)
	if err != nil {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "invalid id provided")
	}

	// Get blob
	blob, err := boiler.FindBlob(passdb.StdConn, blobID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return http.StatusNotFound, terror.Error(err, "attachment not found")
		}
		return http.StatusInternalServerError, terror.Error(err, "could not get attachment")
	}

	// Non-public image? check auth
	if !blob.Public {
		_, code, err := GetUserFromToken(c.API, r)
		if err != nil {
			return code, err
		}
	}

	// Get disposition
	disposition := "attachment"
	isViewDisposition := r.URL.Query().Get("view")
	if isViewDisposition == "true" {
		disposition = "inline"
	}

	// tell the browser the returned content should be downloaded/inline
	if blob.MimeType != "" && blob.MimeType != "unknown" {
		w.Header().Add("Content-Type", blob.MimeType)
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("%s;filename=%s.%s", disposition, blob.FileName, blob.Extension))
	_, err = w.Write(blob.File)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// FileGetByName retrives a file attachment
func (c *FilesController) FileGetByName(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()

	// Get blob id
	fileName := chi.URLParam(r, "name")
	if fileName == "" {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "no name provided")
	}

	// Get blob
	blob, err := boiler.Blobs(
		boiler.BlobWhere.FileName.EQ(fileName),
	).One(passdb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return http.StatusNotFound, terror.Error(err, "attachment not found")
		}
		return http.StatusInternalServerError, terror.Error(err, "could not get attachment")
	}

	// Non-public image? check auth
	if !blob.Public {
		_, code, err := GetUserFromToken(c.API, r)
		if err != nil {
			return code, err
		}
	}

	// Get disposition
	disposition := "attachment"
	isViewDisposition := r.URL.Query().Get("view")
	if isViewDisposition == "true" {
		disposition = "inline"
	}

	// tell the browser the returned content should be downloaded/inline
	if blob.MimeType != "" && blob.MimeType != "unknown" {
		w.Header().Add("Content-Type", blob.MimeType)
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("%s;filename=%s.%s", disposition, blob.FileName, blob.Extension))
	_, err = w.Write(blob.File)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// FileUpload retrives a file attachment
func (c *FilesController) FileUpload(w http.ResponseWriter, r *http.Request, user *boiler.User) (int, error) {
	defer r.Body.Close()

	// Get blob
	blob, _, err := parseUploadRequest(w, r, nil)
	if errors.Is(err, ErrUnknownFileType) {
		return http.StatusBadRequest, terror.Error(err, "file type not allowed")
	}
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "something went wrong, please try again")
	}

	if blob == nil {
		return http.StatusInternalServerError, terror.Error(err, "file is required")
	}

	// Get arguments
	public := r.URL.Query().Get("public")
	if public == "true" {
		blob.Public = true
	}

	// File with the same size and hash exists? return that
	if blob.Hash != nil {
		existingBlob, err := boiler.Blobs(
			boiler.BlobWhere.Hash.EQ(null.StringFromPtr(blob.Hash)),
			boiler.BlobWhere.FileSizeBytes.EQ(blob.FileSizeBytes),
		).One(passdb.StdConn)

		if err == nil {
			if existingBlob != nil && !existingBlob.Public && blob.Public {
				// Make existing blob public
				existingBlob.Public = true
				_, err = existingBlob.Update(passdb.StdConn, boil.Whitelist(boiler.BlobColumns.Public, boiler.BlobColumns.FileName))
				if err != nil {
					return http.StatusInternalServerError, terror.Error(err, "failed to upload")
				}
			}

			// Return existing blob
			return helpers.EncodeJSON(w, struct {
				ID string `json:"id"`
			}{ID: existingBlob.ID})
		}
	}

	// Insert blob
	bb := boiler.Blob{
		FileName:      blob.FileName,
		MimeType:      blob.MimeType,
		FileSizeBytes: blob.FileSizeBytes,
		Extension:     blob.Extension,
		File:          blob.File,
		Hash:          null.StringFromPtr(blob.Hash),
		Public:        blob.Public,
	}
	err = bb.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "failed to upload")
	}

	return helpers.EncodeJSON(w, struct {
		ID types.BlobID `json:"id"`
	}{ID: blob.ID})
}

// parseUploadRequest will read a multipart form request that includes both a file, and a request body
// returns a blob struct, ready to be inserted, as well as decoding json into supplied interface when present
func parseUploadRequest(w http.ResponseWriter, r *http.Request, req interface{}) (*types.Blob, map[string]string, error) {
	// Limit size to 50MB (50<<20)
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	mr, err := r.MultipartReader()
	if err != nil {
		return nil, nil, terror.Error(err, "parse error")
	}

	var blob *types.Blob
	params := make(map[string]string)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, terror.Error(err, "parse error")
		}

		data, err := ioutil.ReadAll(part)
		if err != nil {
			return nil, nil, terror.Error(terror.ErrParse, "parse error")
		}

		// handle file
		if part.FormName() == "file" {
			// get mime type
			kind, err := filetype.Match(data)
			if err != nil {
				return nil, nil, terror.Error(terror.ErrParse, "parse error")
			}

			if kind == filetype.Unknown {
				return nil, nil, terror.Error(ErrUnknownFileType, "")
			}

			mimeType := kind.MIME.Value
			extension := kind.Extension

			hasher := md5.New()
			_, err = hasher.Write(data)
			if err != nil {
				return nil, nil, terror.Error(err, "hash error")
			}
			hashResult := hasher.Sum(nil)
			hash := hex.EncodeToString(hashResult)

			blob = &types.Blob{
				FileName:      part.FileName(),
				MimeType:      mimeType,
				Extension:     extension,
				FileSizeBytes: int64(len(data)),
				File:          data,
				Hash:          &hash,
			}
		} else {
			params[part.FormName()] = string(data)
		}

		// handle JSON body
		if req != nil && part.FormName() == "json" {
			err = json.NewDecoder(part).Decode(req)
			if err != nil {
				return nil, nil, terror.Error(err, "parse error")
			}

		}
	}

	return blob, params, nil
}
