package config

import (
	"os"
	"path"
)

func GetUIDistDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path.Join(dir, "dashboard/build")
}
