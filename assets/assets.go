package assets

import (
	"embed"
	"errors"
	"io/fs"
	"os"
)

var external_fs fs.FS

//go:embed html/*
var embedded_fs embed.FS

func ReadBytes(name string) ([]byte, error) {
	if external_fs != nil {
		b, err := fs.ReadFile(external_fs, name)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			return b, nil
		}
	}
	return fs.ReadFile(embedded_fs, name)
}

func ReadString(name string) (string, error) {
	b, err := ReadBytes(name)
	if b == nil {
		return "", err
	}
	return string(b), err
}

func Stat(name string) (fs.FileInfo, error) {
	if external_fs != nil {
		s, err := fs.Stat(external_fs, name)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			return s, nil
		}
	}
	return fs.Stat(embedded_fs, name)
}

func ReadDir(name string) ([]fs.DirEntry, error) {
	if external_fs != nil {
		s, err := fs.ReadDir(external_fs, name)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			return s, nil
		}
	}
	return fs.ReadDir(embedded_fs, name)
}

func SetExternal(path string) {
	external_fs = os.DirFS(path)
}
