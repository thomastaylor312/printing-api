package store

import (
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

type DiskImageStore struct {
	rootPath string
}

// NewDiskImageStore creates a new DiskImageStore that stores images in the given rootPath. If you
// want to use this to actually serve files, this should be a path you expose as part of your server
func NewDiskImageStore(rootPath string) *DiskImageStore {
	return &DiskImageStore{rootPath: rootPath}
}

func (d *DiskImageStore) Get(container string, id string) (*url.URL, error) {
	fileLocation := filepath.Join(d.rootPath, container, id)
	if _, err := os.Stat(fileLocation); errors.Is(err, os.ErrNotExist) {
		return nil, ErrImageNotFound
	} else if err != nil {
		return nil, err
	}
	return &url.URL{Path: url.PathEscape(fileLocation)}, nil
}

func (d *DiskImageStore) Set(container string, id string, expected_length uint, value io.ReadCloser) (*url.URL, error) {
	fileLocation := filepath.Join(d.rootPath, container, id)
	file, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(file, value)
	// TODO: Validate that this is an image
	return &url.URL{Path: url.PathEscape(fileLocation)}, err
}

func (d *DiskImageStore) Delete(container string, id string) error {
	err := os.Remove(filepath.Join(d.rootPath, container, id))
	if errors.Is(err, os.ErrNotExist) {
		return ErrImageNotFound
	} else if err != nil {
		return err
	}
	return nil
}
