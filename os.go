package main

import (
	"os"
	"path/filepath"
)

// Collect files to add them later
func Walk(dir string) (accessible, unReadable, gtSize []string, symlinks [][2]string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			if info.Mode()&os.ModeSymlink != 0 {
				dest, err := os.Readlink(path)
				if err != nil {
					return err
				}
				symlinks = append(symlinks, [2]string{path, dest})
			} else if info.Mode()&(1<<2) == 0 {
				unReadable = append(unReadable, path)
			} else if info.Size() > 1*1024*1024 {
				gtSize = append(gtSize, path)
			} else {
				accessible = append(accessible, path)
			}
		}
		return err
	})

	return accessible, unReadable, gtSize, symlinks, err
}
