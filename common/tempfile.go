package common

import (
	"fmt"
	"io"
	"os"
	"time"
)

func scheduleRemove(path string) {
	time.Sleep(60 * time.Second)
	os.Remove(path)
}

func SaveTemporaryFile(src io.Reader, name string) {
	path := fmt.Sprintf("public/%s", name)
	dest, err := os.Create(path)
	if err != nil {
		return
	}

	io.Copy(dest, src)
	go scheduleRemove(path)
}
