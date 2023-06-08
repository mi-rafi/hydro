package internal

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io/fs"
	"os"
	"time"
)

type TimeLoader interface {
	GetStartupTime() (time.Time, error)
	StoreStartupTime(time.Time) error
}

type FileTimeLoader struct {
	f string
}

type FileTimeLoaderConfig struct {
	FileName string
}

func NewFileTimeLoader(cfg *FileTimeLoaderConfig) (*FileTimeLoader, func(), error) {
	var _, err = os.Stat(cfg.FileName)
	var file os.File
	if os.IsNotExist(err) {
		file, err := os.Create(cfg.FileName)
		if err != nil {
			fmt.Println(err)

		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Error().Err(err).Msg("can not close data file")
			}
		}(file)
	}
	return &FileTimeLoader{f: cfg.FileName}, func() {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msg("can not close data file")
		}
	}, nil
}

func (f *FileTimeLoader) GetStartupTime() (time.Time, error) {
	data, err := os.ReadFile(f.f)
	if err != nil {
		return time.Time{}, err
	}
	pt := time.Time{}
	err = pt.UnmarshalBinary(data)
	return pt, err
}

func (f *FileTimeLoader) StoreStartupTime(t time.Time) error {
	d, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return os.WriteFile(f.f, d, fs.ModePerm)
}
