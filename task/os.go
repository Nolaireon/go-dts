package task

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	arch        = runtime.GOOS
	symlinkName = "current"
	versionsDir = "versions"
	//excludedAppsConfigName = "excluded_apps.yaml"
)

var (
	ErrUnsupportedOS         = errors.New("unsupported os, go-dts can be run on linux or windows")
	ErrWorkTreeIsAFile       = errors.New("work-tree couldn't be a file")
	ErrVersionLinkIsNotALink = errors.New("version link is not a symlink")
	//ErrOSNotSupportSymlinks = errors.New("os do not support symlinks")
)

// walk recursively walks through the specified directory, distributing each file according to the Files structure
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
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
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

// openLogFile safely open a file at a given path, depend on the using OS the error of nonexistent path can be handled
// in a two different way. Since windows is not supporting symlinks, the only difference is
// that directory will be created instead.
func openLogFile(fPath string) (file *os.File, err error) {
	path := filepath.Dir(fPath)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if arch == "linux" {
			err = os.MkdirAll(logDir, 0755)
			if err != nil {
				return
			}
			err = createSymlink(logDir, path)
			if err != nil {
				return
			}
		} else if arch == "windows" {
			if err := os.MkdirAll(path, 0755); err != nil {
				return file, err
			}
		} else {
			err = ErrUnsupportedOS
			return
		}
	}

	file, err = os.OpenFile(fPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	return
}

// replaceEnv read config from given path and unmarshall data to *State
func (st *State) replaceEnv(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, st)
	if err != nil {
		return err
	}

	return nil
}

func getExcludedApps(path string) ([]string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ea := &ExcludedApps{}

	err = yaml.Unmarshal(b, ea)
	if err != nil {
		return nil, err
	}

	return ea.ExcludedApps, nil
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

// Wrapper on os.Symlink function
func createSymlink(dst, src string) error {
	return os.Symlink(dst, src)
}

// Wrapper on os.Executable function
func getExecutablePath() (path string, err error) {
	path, err = os.Executable()
	if err != nil {

	}

	path = filepath.Dir(path)
	return
}

// Wrapper on filepath.Join function
func joinPaths(elem ...string) string {
	return filepath.Join(elem...)
}

// decomposeWorkTree decompose given work-tree on composite parts
func (env *Environment) decomposeWorkTree(workTree string) (err error) {
	workTree = filepath.Clean(workTree)
	var fi os.FileInfo
	fi, err = os.Lstat(workTree)

	if err != nil {
		return
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		if env.WorkTree, err = os.Readlink(workTree); err != nil {
			return
		}

		env.AppDir = filepath.Dir(workTree)
		env.Instance = getInstance(env.AppDir)
	} else if fi.IsDir() {
		env.WorkTree = workTree
		env.AppDir = workTree
		env.Instance = getInstance(env.AppDir)
	} else {
		err = ErrWorkTreeIsAFile
	}

	return
}

// resolveCurrentVersion возвращает путь до workTree актуальной версии приложения(вычисляется по симлинку),
// если не существует симлинка или приложение не поддерживает версионность, то возвращается workTree = appDir
func resolveCurrentVersion(appDir string) (workTree string, err error) {
	versionsPath := joinPaths(appDir, versionsDir)
	_, err = os.Stat(versionsPath)

	if arch == "windows" || os.IsNotExist(err) {
		return appDir, nil
	}

	// code below executed when os is linux and target app is versioning
	var symlink os.FileInfo
	symlinkPath := joinPaths(appDir, symlinkName)
	symlink, err = os.Lstat(symlinkPath)
	if err != nil {
		return
	}

	if symlink.Mode()&os.ModeSymlink == 0 {
		err = ErrVersionLinkIsNotALink
	}

	workTree, err = os.Readlink(symlinkPath)

	return
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

func removeGitDir(gitDir string) error {
	return os.RemoveAll(gitDir)
}
