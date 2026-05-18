package fsx

import (
	"io/fs"

	"gopkg.in/yaml.v3"
	"os"
)

func ReadYAML(path string, out any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return yaml.Unmarshal(b, out)
}
func ReadYAMLFromFS(fsys fs.FS, path string, out any) error {
	b, e := fs.ReadFile(fsys, path)
	if e != nil {
		return e
	}
	return yaml.Unmarshal(b, out)
}

func WriteYAML(path string, in any) error {
	b, e := yaml.Marshal(in)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0o644)
}
