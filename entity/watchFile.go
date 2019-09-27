package entity

import (
	"github.com/fsnotify/fsnotify"
	. "logDog/common"
	"os"
)

type FileWatcher struct {
	Path    string
	File    *os.File
	Offset  int64
	Watcher *fsnotify.Watcher
}

func NewFileWatcher(path string) *FileWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Logger.Error(err)
		return nil
	}
	err = watcher.Add(path)
	if err != nil {
		Logger.Error(err)
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		Logger.Error(err)
		return nil
	}
	wf := FileWatcher{
		Path:    path,
		File:    file,
		Offset:  0,
		Watcher: watcher,
	}
	return &wf
}

func (wf *FileWatcher) ReWatch() error {
	err := wf.Watcher.Remove(wf.Path)
	if err != nil {
		return err
	}
	err = wf.File.Close()
	if err != nil {
		return err
	}
	wf.File, err = os.Open(wf.Path)
	if err != nil {
		return err
	}
	err = wf.Watcher.Add(wf.Path)
	if err != nil {
		return err
	}
	wf.Offset = 0
	return nil
}
