package unity

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/itchio/lzma"
)

// https://docs.unity3d.com/530/Documentation/Manual/AssetBundleInternalStructure.html
// https://docs.unity3d.com/Manual/LoadingResourcesatRuntime.html

type Asset struct {
	Reader         *Reader
	Bundle         *Bundle
	BundleOffset   int64
	Name           string
	Tree           *TypeMetadata
	Types          map[int32]TypeTree
	Objects        map[int64]ObjectInfo
	Adds           map[int64]int32
	AssetRefs      []*AssetRef
	IsLittleEndian bool
	IsLoaded       bool
	LongObjectIDs  bool
	MetadataSize   uint32
	FileSize       uint32
	Format         uint32
	DataOffset     uint32
}

func AssetFromBundle(bundle *Bundle, reader *Reader) *Asset {
	a := Asset{
		Bundle:  bundle,
		Adds:    make(map[int64]int32),
		Objects: make(map[int64]ObjectInfo),
		Types:   make(map[int32]TypeTree),
	}

	offset := reader.Tell()

	if bundle.Signature == SignatureFS {
		a.Reader = reader
		a.BundleOffset = offset
		return &a
	}

	if bundle.Compressed() {
		ofs := reader.Tell()
		raw, _ := reader.Bytes(reader.Remaining())
		decompressed := lzma.NewReader(bytes.NewReader(raw))
		b, err := ioutil.ReadAll(decompressed)
		if err != nil {
			return nil
		}
		a.Reader, _ = NewReader(b)
		a.BundleOffset = 0
		reader.SeekStart(ofs)
	} else {
		a.Name, _ = reader.StringNull()
		headerSize, _ := reader.Uint32()
		_, _ = reader.Uint32()
		a.BundleOffset = offset + int64(headerSize) - 4
		if a.IsResource() {
			a.BundleOffset -= int64(len(a.Name))
		}
		a.Reader = reader
	}

	return &a
}

//func AssetFromFile()

func (a *Asset) String() string {
	return fmt.Sprintf("%#v", a)
}

func (a *Asset) ReadID(reader *Reader) (int64, error) {
	if a.Format >= 14 {
		return reader.Int64()
	}

	if id, err := reader.Int32(); err != nil {
		return 0, err
	} else {
		return int64(id), nil
	}
}

func (a *Asset) IsResource() bool {
	return strings.HasSuffix(a.Name, ".resource")
}

func (a *Asset) LoadObjects(sig string) error {
	if !a.IsLoaded {
		return a.Load(sig)
	}

	return nil
}

func (a *Asset) Load(sig string) error {
	if a.IsResource() {
		a.IsLoaded = true
		return nil
	}

	if sig == SignatureFS || sig == SignatureRaw {
		return a.LoadFromBuffer()
	} else {
		return fmt.Errorf("unity.Asset.Load: Signature not supported: %v", sig)
	}
}

func (a *Asset) LoadFromBuffer() (err error) {
	if _, err := a.Reader.SeekStart(a.BundleOffset); err != nil {
		return fmt.Errorf("unity.Asset.LoadFromBuffer: Couldn't seek to offset %v", a.BundleOffset)
	}
	a.MetadataSize, err = a.Reader.Uint32()
	a.FileSize, _ = a.Reader.Uint32()
	a.Format, _ = a.Reader.Uint32()
	a.DataOffset, _ = a.Reader.Uint32()

	if a.Format >= 9 {
		endian, _ := a.Reader.Uint32()
		a.IsLittleEndian = endian == 0
		a.Reader.ChangeEndian(a.IsLittleEndian)
	}

	if a.Tree, err = ReadTypeMetadata(a.Reader, a.IsLittleEndian, a.Format); err != nil {
		return err
	}

	if a.Format >= 7 && a.Format < 14 {
		longObjectIds, _ := a.Reader.Uint32()
		a.LongObjectIDs = longObjectIds != 0
	}

	numObjects, _ := a.Reader.Uint32()

	for i := uint32(0); i < numObjects; i++ {
		if a.Format >= 14 {
			a.Reader.Align()
		}

		obj, err := ReadObjectInfo(a, a.Reader)
		if err != nil {
			return err
		}

		a.registerObject(obj)
	}

	if a.Format >= 11 {
		numAdds, _ := a.Reader.Uint32()
		for i := uint32(0); i < numAdds; i++ {
			if a.Format >= 14 {
				a.Reader.Align()
			}

			id, _ := a.ReadID(a.Reader)
			add, _ := a.Reader.Int32()
			a.Adds[id] = add
		}
	}

	if a.Format >= 6 {
		numRefs, _ := a.Reader.Uint32()
		for i := uint32(0); i < numRefs; i++ {
			ref, err := ReadAssetRef(a.Reader, a.Format, a.IsLittleEndian)
			if err != nil {
				return err
			}
			a.AssetRefs = append(a.AssetRefs, ref)
		}
	}

	if unknownStr, err := a.Reader.StringNull(); err != nil {
		return err
	} else if unknownStr != "" {
		return errors.New("unity.Asset.LoadFromBuffer: Ending string not empty.")
	}

	a.IsLoaded = true

	return nil
}

func (a *Asset) registerObject(obj *ObjectInfo) error {
	if tree, found := a.Tree.TypeTrees[obj.TypeID]; found {
		a.Types[obj.TypeID] = tree
	} else if tree, found = a.Types[obj.TypeID]; !found {
		// TODO: default metadata
	}

	if _, found := a.Objects[obj.PathID]; found {
		return fmt.Errorf("unity.Asset.registerObject: Duplicate asset object: %v", a.Objects[obj.PathID])
	}

	a.Objects[obj.PathID] = *obj

	return nil
}
