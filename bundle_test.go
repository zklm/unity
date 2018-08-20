package unity

import (
	"path/filepath"
	"testing"
)

func TestReadBundle(t *testing.T) {
	path, _ := filepath.Abs("test/20147_cs_h")
	bundle, err := ReadBundle(path)
	if err != nil {
		t.Error(err.Error())
	}

	if err = bundle.ResolveAsset(0); err != nil {
		t.Error(err.Error())
	}

	asset := bundle.Assets[0]
	if asset.Name != "CAB-be1d08a614f11a49e601c02ba4c4f640" {
		t.Error(err.Error())
	}

	objects := asset.Objects
	if len(objects) != 2 {
		t.Error(err.Error())
	}
}
