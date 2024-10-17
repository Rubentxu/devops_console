package config

import (
    "os"
)

func IsDevelopmentMode() bool {
    return os.Getenv("DEV_MODE") == "true"
}