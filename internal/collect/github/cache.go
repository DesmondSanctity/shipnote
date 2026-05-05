package github

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrCacheMiss is returned by Cache.Get when no entry is stored for the key.
var ErrCacheMiss = errors.New("github cache: miss")

// Cache stores and retrieves opaque response bodies keyed by an arbitrary
// string. Implementations must be safe for concurrent use.
type Cache interface {
	Get(key string) ([]byte, error)
	Put(key string, body []byte) error
}

// FileCache is a Cache backed by a directory on disk. Each key is hashed
// to a 64-char hex SHA-256 and stored as a single file. Entries are
// immutable; replacing a key writes a new file atomically via rename.
//
// The cache is content-addressed: callers should compute the key from
// the inputs that fully determine the response (e.g. query text + JSON
// variables). Two identical inputs always hit the same file.
type FileCache struct {
	dir string
}

// NewFileCache returns a FileCache rooted at dir. The directory is
// created on first write; reads from a missing directory simply miss.
func NewFileCache(dir string) *FileCache {
	return &FileCache{dir: filepath.Clean(dir)}
}

// HashKey turns arbitrary input into a stable filename-safe key.
// Callers can also pass an already-hashed key; HashKey is idempotent on
// 64-char hex strings.
func HashKey(input string) string {
	if len(input) == 64 && isHex(input) {
		return input
	}
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

// Get returns the cached body for key, or ErrCacheMiss when absent.
func (c *FileCache) Get(key string) ([]byte, error) {
	b, err := os.ReadFile(c.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrCacheMiss
	}
	return b, err
}

// Put stores body under key, atomically replacing any prior entry.
func (c *FileCache) Put(key string, body []byte) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	final := c.path(key)
	tmp, err := os.CreateTemp(c.dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, final)
}

func (c *FileCache) path(key string) string {
	return filepath.Join(c.dir, HashKey(key))
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

// nopCache is the no-op Cache used when the user has not configured one.
type nopCache struct{}

func (nopCache) Get(string) ([]byte, error) { return nil, ErrCacheMiss }
func (nopCache) Put(string, []byte) error   { return nil }

var _ Cache = nopCache{}

// keyFor builds the cache key for a GraphQL request. Two requests with
// identical query text and variables JSON share the same key.
func keyFor(query, varsJSON string) string {
	return HashKey(strings.Join([]string{query, varsJSON}, "\x00"))
}
