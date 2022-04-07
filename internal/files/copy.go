package files

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyItem copies a file or directory from src to dst.
func CopyItem(dst, src string, force bool) error {
	if st, err := os.Stat(src); err == nil && st.IsDir() {
		return CopyDir(dst, src, force)
	} else if err == nil {
		return CopyFile(dst, src, force)
	} else {
		return err
	}
}

// CopyDir recursively copies a directory from src to dst.
func CopyDir(dst, src string, force bool) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return CopyFile(filepath.Join(dst, name), path, force)
	})
}

// CopyFile copies a file from src to dst. If there are missing directories in
// dst, they are created.
func CopyFile(dst, src string, force bool) error {
	srcf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcf.Close()
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !force {
		flags = flags | os.O_EXCL
	}

	dstDir := filepath.Dir(dst)
	if dstDir != "." {
		err = os.MkdirAll(dstDir, 0777) // leave it to umask
		if err != nil {
			return err
		}
	}

	dstf, err := os.OpenFile(dst, flags, 0666) // leave it to umask
	if err != nil {
		return err
	}
	defer dstf.Close()

	_, err = io.Copy(dstf, srcf)
	if err != nil {
		return err
	}
	return nil
}
