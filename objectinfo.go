package unity

import (
	"fmt"
)

type ObjectInfo struct {
	PathID     int64
	DataOffset uint32
	Size       uint32
	TypeID     int32
	ClassID    int16
	// Not needed
	// IsDestroyed bool  `version:"<11"`
	// Unknown0    int16 `version:">=11"`
	// Unknown1    byte  `version:">=15"`
}

func ReadObjectInfo(asset *Asset, reader *Reader) (obj *ObjectInfo, err error) {
	obj = &ObjectInfo{}
	if obj.PathID, err = obj.readID(asset, reader); err != nil {
		return
	}

	if obj.DataOffset, err = reader.Uint32(); err != nil {
		return
	}

	if obj.Size, err = reader.Uint32(); err != nil {
		return
	}

	if asset.Format < 17 {
		if obj.TypeID, err = reader.Int32(); err != nil {
			return
		}

		if obj.ClassID, err = reader.Int16(); err != nil {
			return
		}
	} else {
		typeIndex, _ := reader.Int32()
		if len(asset.Tree.ClassIDs) <= int(typeIndex) {
			return nil, fmt.Errorf("unity.ReadObjectInfo: Undefined type metadata. TypeIndex: %v", typeIndex)
		}
		classID := asset.Tree.ClassIDs[typeIndex]
		obj.TypeID = classID
		obj.ClassID = int16(classID)
	}

	if asset.Format < 11 {
		if _, err = reader.Int16(); err != nil {
			return
		}
	} else if asset.Format >= 11 && asset.Format < 17 {
		if _, err = reader.Int16(); err != nil {
			return
		}

		if asset.Format >= 15 {
			if _, err = reader.Int8(); err != nil {
				return
			}
		}
	}

	return
}

func (oi *ObjectInfo) String() string {
	return fmt.Sprintf("Type: %v, Path: %v, Class: %v, Size: %v", oi.TypeID, oi.PathID, oi.ClassID, oi.Size)
}

func (oi *ObjectInfo) readID(asset *Asset, reader *Reader) (int64, error) {
	if asset.LongObjectIDs {
		return reader.Int64()
	}
	return asset.ReadID(reader)
}
