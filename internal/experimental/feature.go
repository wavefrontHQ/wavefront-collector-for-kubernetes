package experimental

import "os"

func IsEnabled(name string) bool {
    return len(os.Getenv(name)) > 0
}

func EnableFeature(name string) {
    os.Setenv(name, name)
}