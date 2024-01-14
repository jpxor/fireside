package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	pathUtil "path"
	"path/filepath"
	"strings"
)

var root string

func InitUserFilesRootDir(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(absPath, 0660)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	root = absPath
}

func DirectoryListing(uid, relpath string) ([]fs.DirEntry, error) {
	// don't let relpath break out of the app' root path
	absPath := pathUtil.Clean(
		filepath.Join(root, uid, relpath),
	)
	if !strings.HasPrefix(absPath, root) {
		return nil, fmt.Errorf("invalid path")
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(absPath, 0660)
			if err == nil {
				// try again
				return DirectoryListing(uid, relpath)
			}
		}
		return nil, fmt.Errorf("unknown path")
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}
	return os.ReadDir(absPath)
}

func CreateFolder(uid, relpath, name string) error {
	// don't let relpath break out of this context's working directory
	absFilepath := pathUtil.Clean(
		filepath.Join(root, uid, relpath, name),
	)
	if !strings.HasPrefix(absFilepath, root) {
		return fmt.Errorf("invalid path")
	}
	_, err := os.Stat(absFilepath)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file with that name already exists")
	}
	return os.MkdirAll(absFilepath, 0660)
}

func CreateJournal(uid, relpath, name string) error {
	// don't let relPath break out of this context's working directory
	absFilepath := pathUtil.Clean(
		filepath.Join(root, uid, relpath, name),
	)
	if !strings.HasPrefix(absFilepath, root) {
		return fmt.Errorf("invalid path")
	}
	if !strings.HasSuffix(absFilepath, ".journal") {
		absFilepath = absFilepath + ".journal"
	}
	_, err := os.Stat(absFilepath)
	if err == nil {
		return fmt.Errorf("file with that name already exists")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("invalid filename")
	}
	f, err := os.OpenFile(absFilepath, os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("failed to create file")
	}
	f.Close()
	return nil
}
