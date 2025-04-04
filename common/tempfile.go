package common

import (
	"fmt"
	"io"
	"os"
	"time"
)

func scheduleRemove(name string) {
	time.Sleep(60 * time.Second)
	DeletePublicFile(name)
}

func DeletePublicFile(name string) error {
	path := fmt.Sprintf("public/%s", name)
	return os.Remove(path)
}

func SavePublicFile(src io.Reader, name string) error {
	path := fmt.Sprintf("public/%s", name)
	dest, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(dest, src)
	return err
}

func SaveTemporaryFile(src io.Reader, name string) error {
	err := SavePublicFile(src, name)
	if err != nil {
		return err
	}

	go scheduleRemove(name)
	return nil
}
