package unity

type AssetRef struct {
	AssetPath string
	GUID      []byte
	Type      int32
	FilePath  string
}

func ReadAssetRef(reader *Reader, format uint32, isLittleEndian bool) (ref *AssetRef, err error) {
	ref = &AssetRef{}

	if ref.AssetPath, err = reader.StringNull(); err != nil {
		return
	}

	if ref.GUID, err = reader.Bytes(16); err != nil {
		return
	}

	if ref.Type, err = reader.Int32(); err != nil {
		return
	}

	if ref.FilePath, err = reader.StringNull(); err != nil {
		return
	}

	return
}
