package metadata

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Repo struct {
	Root             string
	MetadataFilename string
}

func (r *Repo) MetadataFiles() ([]MetadataFile, error) {
	files := make([]MetadataFile, 0)
	err := filepath.WalkDir(r.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() == r.MetadataFilename {
			f, err := r.newFile(path)
			if err != nil {
				return err
			}
			files = append(files, f)
		}

		return nil
	})
	return files, err
}

func (r *Repo) ReadFile(pathRelativeToRoot string) (string, error) {
	fullPath := filepath.Join(r.Root, pathRelativeToRoot)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("Could not get contents of %s: %v", fullPath, err)
	}
	return string(content), nil
}

func (r *Repo) newFile(fullPath string) (MetadataFile, error) {
	relativePath, err := filepath.Rel(r.Root, fullPath)
	if err != nil {
		return MetadataFile{}, err
	}
	return MetadataFile{
		path:               fullPath,
		pathRelativeToRoot: relativePath,
		repo:               r,
	}, nil

}

type MetadataFile struct {
	path               string
	pathRelativeToRoot string
	repo               *Repo
}

func (f *MetadataFile) Contents() (string, error) {
	return f.repo.ReadFile(f.pathRelativeToRoot)
}

func (f *MetadataFile) Dirs() []string {
	dir := filepath.Dir(f.pathRelativeToRoot)
	if dir == "." {
		return []string{}
	}

	return strings.Split(dir, string(filepath.Separator))
}
