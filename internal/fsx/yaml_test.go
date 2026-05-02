package fsx

import (
	"path/filepath"
	"testing"
)

type d struct{ A string `yaml:"a"` }

func TestYAML(t *testing.T){ p:=filepath.Join(t.TempDir(),"a.yaml"); if err:=WriteYAML(p,d{A:"x"}); err!=nil { t.Fatal(err)}; var out d; if err:=ReadYAML(p,&out); err!=nil { t.Fatal(err)}; if out.A!="x" { t.Fatal("mismatch") } }
