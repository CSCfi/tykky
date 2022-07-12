package tykky

import (
	"os"
	"path/filepath"
)

type wrapperInfo struct {
	location string
	name     string
}

func createWrapper() wrapperInfo {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return wrapperInfo{filepath.Dir(ex), filepath.Base(ex)}
}
