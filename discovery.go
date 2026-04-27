package main

import (
	"io/fs"
	"os"
	"path/filepath"
)

// DiscoverEPWINGBooks walks root recursively and returns the path of every
// directory that appears to be an EPWING book root.
//
// A directory is treated as an EPWING book root when it contains a CATALOG
// or CATALOGS file (in any combination of upper/lower case).  EPWING was
// designed for case-insensitive file systems so both casings are checked.
//
// Once a book root is identified its subdirectories are not visited, because
// they belong to that book rather than being independent books.
//
// Paths that cannot be accessed are silently skipped.
func DiscoverEPWINGBooks(root string) ([]string, error) {
	var books []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible paths silently.
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if isEPWINGBookRoot(path) {
			books = append(books, path)
			return filepath.SkipDir // do not recurse into the book's own subdirectories
		}
		return nil
	})
	return books, err
}

// isEPWINGBookRoot returns true when dir contains a CATALOG or CATALOGS file.
// The presence of either file is the standard indicator of an EPWING book root.
func isEPWINGBookRoot(dir string) bool {
	for _, name := range []string{"CATALOG", "catalog", "CATALOGS", "catalogs", "Catalog", "Catalogs"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}
