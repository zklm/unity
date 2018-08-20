package unity

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestReadBundle(t *testing.T) {
	bundles := []struct {
		path      string
		name      string
		objectLen int
	}{
		{"test/20147_cs_h", "CAB-be1d08a614f11a49e601c02ba4c4f640", 2},
		{"test/main_dxt1_bc1.unity3d", "CAB-ba01e3c16ba268ec36e9543a39dc83ad", 4},
	}

	for _, tc := range bundles {
		path, _ := filepath.Abs(tc.path)
		bundle, err := ReadBundle(path)
		if err != nil {
			t.Error(err.Error())
		}

		if err = bundle.ResolveAsset(0); err != nil {
			t.Error(err.Error())
		}

		asset := bundle.Assets[0]
		if asset.Name != tc.name {
			t.Error(fmt.Errorf("Invalid asset name. Got: %s Expected: %s", asset.Name, tc.name))
		}

		objects := asset.Objects
		if len(objects) != tc.objectLen {
			t.Error(fmt.Errorf("Invalid object count. Got: %v Expected: %v", len(objects), tc.objectLen))
		}
	}
}
