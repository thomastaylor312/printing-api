package store

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type DiskImageStore struct {
	rootPath string
}

func NewDiskImageStore(rootPath string) *DiskImageStore {
	return &DiskImageStore{rootPath: rootPath}
}

func (d *DiskImageStore) Get(container string, id string) (io.ReadCloser, error) {
	file, err := os.Open(filepath.Join(d.rootPath, container, id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrImageNotFound
	} else if err != nil {
		return nil, err
	}
	return file, nil
}

func (d *DiskImageStore) Set(container string, id string, value io.ReadCloser) error {
	file, err := os.OpenFile(filepath.Join(d.rootPath, container, id), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, value)
	return err
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
