package unity

type TypeMetadata struct {
	GeneratorVersion string
	TargetPlatform   uint32
	Hashes           map[int32][]byte
	TypeTrees        map[int32]TypeTree
	ClassIDs         []int32
	HasTypeTrees     bool
	NumTypes         int32
}

func ReadTypeMetadata(reader *Reader, isLittleEndian bool, formatVer uint32) (*TypeMetadata, error) {
	var err error
	tm := TypeMetadata{
		Hashes:    make(map[int32][]byte),
		TypeTrees: make(map[int32]TypeTree),
	}

	if tm.GeneratorVersion, err = reader.StringNull(); err != nil {
		return nil, err
	}

	if tm.TargetPlatform, err = reader.Uint32(); err != nil {
		return nil, err
	}

	if formatVer >= 13 {
		if hasTypeTrees, err := reader.Uint8(); err != nil {
			return nil, err
		} else {
			tm.HasTypeTrees = hasTypeTrees > 0
		}

		if tm.NumTypes, err = reader.Int32(); err != nil {
			return nil, err
		}

		for i := int32(0); i < tm.NumTypes; i++ {
			var hash []byte
			classID, err := reader.Int32()
			if err != nil {
				return nil, err
			}

			if formatVer >= 17 {
				_, _ = reader.Uint8()
				scriptID, _ := reader.Int16()
				if classID == 114 {
					if scriptID >= 0 {
						classID = -2 - int32(scriptID)
					} else {
						classID = -1
					}
				}
			}

			if classID < 0 {
				hash, err = reader.Bytes(0x20)
			} else {
				hash, err = reader.Bytes(0x10)
			}

			if err != nil {
				return nil, err
			}

			tm.ClassIDs = append(tm.ClassIDs, classID)
			tm.Hashes[classID] = hash

			if tm.HasTypeTrees {
				if tt, err := ReadTypeTree(reader, isLittleEndian, formatVer); err != nil {
					return nil, err
				} else {
					tm.TypeTrees[classID] = *tt
				}
			}
		}
	} else {
		numFields, err := reader.Int32()
		if err != nil {
			return nil, err
		}

		for i := int32(0); i < numFields; i++ {
			classID, err := reader.Int32()
			if err != nil {
				return nil, err
			}

			if tt, err := ReadTypeTree(reader, isLittleEndian, formatVer); err != nil {
				return nil, err
			} else {
				tm.TypeTrees[classID] = *tt
			}
		}
	}

	return &tm, nil
}
