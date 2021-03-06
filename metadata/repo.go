package metadata

import (
	"io/fs"
	"path/filepath"
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

func (r *Repo) newFile(fullPath string) (MetadataFile, error) {
	relativePath, err := filepath.Rel(r.Root, fullPath)
	if err != nil {
		return MetadataFile{}, err
	}
	return MetadataFile{
		pathRelativeToRoot: relativePath,
		repo:               r,
	}, nil

}

type MetadataFile struct {
	pathRelativeToRoot string
	repo               *Repo
}
