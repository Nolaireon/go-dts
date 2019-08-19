package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//const (
//	platform = runtime.GOOS
//)

// Collect files to add them later
func (f *Files) walk(dir string) (err error) {
	var accessible, unReadable, gtSize []string
	var symlinks [][2]string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			relPath, _ := filepath.Rel(dir, path)
			//Log.Println(relPath)
			if info.Mode()&os.ModeSymlink != 0 {
				dst, err := os.Readlink(path)
				if err != nil {
					return err
				}

				symlinks = append(symlinks, [2]string{relPath, dst})
			} else if info.Mode()&(1<<2) == 0 {
				unReadable = append(unReadable, relPath)
			} else if info.Size() > 1*1024*1024 {
				gtSize = append(gtSize, relPath)
			} else {
				accessible = append(accessible, relPath)
			}
		} else {
			return err
		}

		return err
	})

	f.Accessible = accessible
	f.GtSize = gtSize
	f.UnReadable = unReadable
	f.Symlinks = symlinks

	return
}

// Open file and avoid path isNotExist error
func openFile(fPath string) (file *os.File, err error) {
	path := filepath.Dir(fPath)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 755); err != nil {
			return file, err
		}
	}

	file, err = os.OpenFile(fPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	return
}

// Rotate log (json or plain)
func rotate(src string) (dst string, err error) {
	ext := filepath.Ext(src)
	extLess := src[:len(src)-len(ext)]
	ctime := time.Now().Format("20060102_1504")
	dst = fmt.Sprintf("%s_%s.%s", extLess, ctime, ext)
	err = mv(src, dst)
	return
}

func mv(src, dst string) error {
	return os.Rename(src, dst)
}

func joinPaths(elem ...string) string {
	return filepath.Join(elem...)
}

// Removing "current" from given path
func leaveCurrent(workTree string) string {
	wt := strings.Split(workTree, "/")
	for i := 0; i < len(wt); i++ {
		if wt[i] == "current" || i == len(wt)-1 && len(wt[i]) == 0 {
			return strings.Join(wt[:i], "/")
		}
	}
	return strings.Join(wt, "/")
}

// GetShortHostName return short name of domain
func getShortHostName() (sName string, err error) {
	hostName, err := os.Hostname()
	if err != nil {
		return
	}

	sName = strings.ReplaceAll(hostName, ".megafon.ru", "")
	return
}
