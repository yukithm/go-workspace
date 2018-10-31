package workspace

import (
	"os"
	"path/filepath"
)

// Workspace is a working directory that contains destination and temporary path.
type Workspace struct {
	Dir        string // Destination directory
	TempDir    string // Temporary directory
	TempSuffix string // Temporary suffix
	Cleanup    bool   // Clean up on error
}

func (w *Workspace) NewDir(name string, callback func(ws *Workspace) error) (string, error) {
	dest := filepath.Clean(filepath.Join(w.Dir, name))
	temp := w.tempPath(name)

	needCleanup := w.Cleanup && !fileExists(temp)
	if err := os.MkdirAll(temp, 0755); err != nil {
		return "", err
	}

	cleanup := func() {
		if needCleanup {
			os.RemoveAll(temp)
		}
	}

	defer func() {
		if err := recover(); err != nil {
			cleanup()
			panic(err) // re-throw
		}
	}()

	// Do not set TempDir for sub workspace. Workspace is already in temp directory.
	sub := &Workspace{
		Dir:        temp,
		TempDir:    "",
		TempSuffix: w.TempSuffix,
		Cleanup:    w.Cleanup,
	}
	if err := callback(sub); err != nil {
		cleanup()
		return "", err
	}
	if fileExists(temp) {
		if err := move(temp, dest); err != nil {
			cleanup()
			return "", err
		}
	}

	return dest, nil
}

func (w *Workspace) NewFile(name string, callback func(path string) error) (string, error) {
	dest := filepath.Clean(filepath.Join(w.Dir, name))
	temp := w.tempPath(name)

	parent := filepath.Dir(temp)
	if parent != "." && parent != string(filepath.Separator) {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return "", err
		}
	}

	cleanup := func() {
		if w.Cleanup {
			os.Remove(temp)
		}
	}

	defer func() {
		if err := recover(); err != nil {
			cleanup()
			panic(err) // re-throw
		}
	}()

	if err := callback(temp); err != nil {
		cleanup()
		return "", err
	}
	if fileExists(temp) {
		if err := move(temp, dest); err != nil {
			cleanup()
			return "", err
		}
	}

	return dest, nil
}

func (w *Workspace) tempPath(name string) string {
	dir := w.TempDir
	if dir == "" {
		dir = w.Dir
	}

	return filepath.Clean(filepath.Join(dir, name+w.TempSuffix))
}

func move(from, to string) error {
	if from == to {
		return nil
	}

	parent := filepath.Dir(to)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	return os.Rename(from, to)
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}
