package repo

import "os"

// getenv is os.Getenv wrapped so tests can stub it if needed and to keep
// the sanitizedEnv list focused on the variables we actually depend on.
func getenv(key string) string { return os.Getenv(key) }
