package envars

import "os"

func SetIfNotExists(key, value string) {
	_, ok := os.LookupEnv(key)
	if !ok {
		os.Setenv(key, value)
	}
}
