package fsx

import (
	"os"
	"gopkg.in/yaml.v3"
)

func ReadYAML(path string, out any) error { b, e := os.ReadFile(path); if e != nil { return e }; return yaml.Unmarshal(b, out) }
func WriteYAML(path string, in any) error { b, e := yaml.Marshal(in); if e != nil { return e }; return os.WriteFile(path, b, 0o644) }
