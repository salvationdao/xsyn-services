package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// BlobGet returns a blob by given ID
func BlobGet(ctx context.Context, conn Conn, blobID types.BlobID) (*types.Blob, error) {
	blob := &types.Blob{}
	q := `
		SELECT id, file_name, file_size_bytes, extension, mime_type, file, views, public
		FROM blobs
		WHERE id = $1`
	err := pgxscan.Get(ctx, conn, blob, q, blobID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return blob, nil
}

// BlobGetByFilename returns a blob by filename
func BlobGetByFilename(ctx context.Context, conn Conn, fileName string) (*types.Blob, error) {
	blob := &types.Blob{}
	q := `
		SELECT id, file_name, file_size_bytes, extension, mime_type, file, views, public
		FROM blobs
		WHERE file_name = $1
		LIMIT 1`
	err := pgxscan.Get(ctx, conn, blob, q, fileName)
	if err != nil {
		return nil, terror.Error(err)
	}
	return blob, nil
}

// BlobGetByHash returns a blob by hash and file size
func BlobGetByHash(ctx context.Context, conn Conn, hash string, fileSizeBytes int64) (*types.Blob, error) {
	blob := &types.Blob{}
	q := `
		SELECT id, file_name, file_size_bytes, extension, mime_type, file, views, public
		FROM blobs
		WHERE file_size_bytes = $1 AND hash = $2`
	err := pgxscan.Get(ctx, conn, blob, q, fileSizeBytes, hash)
	if err != nil {
		return nil, terror.Error(err)
	}
	return blob, nil
}

// BlobInsert inserts a new  blob
func BlobInsert(ctx context.Context, conn Conn, blob *types.Blob) error {
	q := `--sql
		INSERT INTO blobs (file_name, mime_type, file_size_bytes, extension, file, hash, public) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, file_name, mime_type, file_size_bytes, extension, file, hash, public`
	err := pgxscan.Get(
		ctx,
		conn,
		blob,
		q,
		blob.FileName,
		blob.MimeType,
		blob.FileSizeBytes,
		blob.Extension,
		blob.File,
		blob.Hash,
		blob.Public,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// BlobUpdate updates a blob's fileName and public status
func BlobUpdate(ctx context.Context, conn Conn, blob *types.Blob) error {
	q := `--sql
		UPDATE blobs 
		SET file_name = $2, public = $3
		WHERE id = $1`
	_, err := conn.Exec(
		ctx,
		q,
		blob.ID,
		blob.FileName,
		blob.Public,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// BlobList returns a list of blobs in the system (excluding user avatars)
func BlobList(ctx context.Context, conn Conn, search string, offset int, limit int) (int, *[]*types.Blob, error) {
	args := []interface{}{}
	var totalRows int
	result := []*types.Blob{}

	// Setup query
	searchCondition := ""
	if search != "" {
		args = append(args, "%"+search+"%")
		searchCondition = fmt.Sprintf(" AND file_name ILIKE $%d", len(args))
	}

	q := `--sql
	FROM blobs
	LEFT JOIN (SELECT avatar_id FROM users) u ON u.avatar_id = blobs.id
	WHERE u.avatar_id IS NULL AND (mime_type = 'image/jpeg' OR mime_type = 'image/png')
	` + searchCondition

	// Get total
	countQ := `SELECt COUNT(blobs.id) ` + q
	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, &result, nil
	}

	// Limit
	limitCondition := ""
	if limit > 0 {
		limitCondition = fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	// Get records
	q = `SELECT blobs.id, blobs.file_name ` + q + limitCondition
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, nil, terror.Error(err)
	}

	return totalRows, &result, nil
}
