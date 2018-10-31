package workspace_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workspace "github.com/yukithm/go-workspace"
)

const testDir = "./testdir"

var (
	destDir = filepath.Join(testDir, "dest")
	tempDir = filepath.Join(testDir, "tmp")
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func assertPath(t *testing.T, want, got string) {
	if want != got {
		t.Errorf(`want "%s", but "%s"`, want, got)
	}
	assertExists(t, got)
}

func assertExists(t *testing.T, path ...string) {
	p := filepath.Join(path...)
	if !exists(p) {
		t.Errorf(`"%s" not exists`, p)
	}
}

func assertNotExists(t *testing.T, path ...string) {
	p := filepath.Join(path...)
	if exists(p) {
		t.Errorf(`"%s" exists`, p)
	}
}

func beforeEach() {
	if err := os.Mkdir(testDir, 0755); err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func afterEach() {
	if err := os.RemoveAll(testDir); err != nil {
		panic(err)
	}
}

func TestNewDir(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    tempDir,
		TempSuffix: ".tmp",
		Cleanup:    false,
	}

	dir, err := ws.NewDir("test.d", func(sub *workspace.Workspace) error {
		path := filepath.Join(testDir, "tmp", "test.d.tmp")
		assertPath(t, path, sub.Dir)
		return nil
	})
	if err != nil {
		t.Fatal("An error is returned:", err)
	}

	path := filepath.Join(testDir, "dest", "test.d")
	assertPath(t, path, dir)
}

func TestNewDirWithoutTempDir(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    "",
		TempSuffix: ".tmp",
		Cleanup:    false,
	}

	dir, err := ws.NewDir("test.d", func(sub *workspace.Workspace) error {
		path := filepath.Join(testDir, "dest", "test.d.tmp")
		assertPath(t, path, sub.Dir)
		return nil
	})
	if err != nil {
		t.Fatal("An error is returned:", err)
	}

	path := filepath.Join(testDir, "dest", "test.d")
	assertPath(t, path, dir)
}

func TestNewDirWithoutTempSuffix(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    tempDir,
		TempSuffix: "",
		Cleanup:    false,
	}

	dir, err := ws.NewDir("test.d", func(sub *workspace.Workspace) error {
		path := filepath.Join(testDir, "tmp", "test.d")
		assertPath(t, path, sub.Dir)
		return nil
	})
	if err != nil {
		t.Fatal("An error is returned:", err)
	}

	path := filepath.Join(testDir, "dest", "test.d")
	assertPath(t, path, dir)
}

func TestNewFile(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    tempDir,
		TempSuffix: ".tmp",
		Cleanup:    false,
	}

	file, err := ws.NewFile("test.dat", func(file string) error {
		path := filepath.Join(testDir, "tmp", "test.dat.tmp")
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		assertPath(t, path, file)
		return nil
	})
	if err != nil {
		t.Fatal("An error is returned:", err)
	}

	path := filepath.Join(testDir, "dest", "test.dat")
	assertPath(t, path, file)
}

func TestNewFileWithoutTempSuffix(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    tempDir,
		TempSuffix: "",
		Cleanup:    false,
	}

	file, err := ws.NewFile("test.dat", func(file string) error {
		path := filepath.Join(testDir, "tmp", "test.dat")
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		assertPath(t, path, file)
		return nil
	})
	if err != nil {
		t.Fatal("An error is returned:", err)
	}

	path := filepath.Join(testDir, "dest", "test.dat")
	assertPath(t, path, file)
}

func TestNewDirAndFile(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    tempDir,
		TempSuffix: ".tmp",
		Cleanup:    false,
	}

	dir, err := ws.NewDir("test.d", func(sub *workspace.Workspace) error {
		var err error
		_, err = sub.NewFile("test.dat", func(file string) error {
			path := filepath.Join(testDir, "tmp", "test.d.tmp", "test.dat.tmp")
			f, err := os.Create(file)
			if err != nil {
				return err
			}
			defer f.Close()
			assertPath(t, path, file)
			return nil
		})
		if err != nil {
			t.Fatal("An error is returned:", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal("An error is returned:", err)
	}

	path := filepath.Join(testDir, "dest", "test.d")
	assertPath(t, path, dir)
	assertExists(t, testDir, "dest", "test.d", "test.dat")
}

func TestCleanup(t *testing.T) {
	beforeEach()
	defer afterEach()

	ws := workspace.Workspace{
		Dir:        destDir,
		TempDir:    tempDir,
		TempSuffix: ".tmp",
		Cleanup:    true,
	}

	_, err := ws.NewDir("test.d", func(sub *workspace.Workspace) error {
		_, err := sub.NewFile("test.dat", func(file string) error {
			f, err := os.Create(file)
			if err != nil {
				return err
			}
			defer f.Close()
			return errors.New("test error")
		})

		if err == nil {
			t.Error("want err, but nil")
		} else if err.Error() != "test error" {
			t.Errorf(`want "test error", but "%s"`, err.Error())
		}

		assertNotExists(t, testDir, "tmp", "test.d.tmp", "test.dat.tmp")
		assertNotExists(t, testDir, "tmp", "test.d.tmp", "test.dat")
		return err
	})

	if err == nil {
		t.Error("want err, but nil")
	} else if err.Error() != "test error" {
		t.Errorf(`want "test error", but "%s"`, err.Error())
	}

	assertNotExists(t, testDir, "tmp", "test.d.tmp")
	assertNotExists(t, testDir, "dest", "test.d")
}
