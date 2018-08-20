package unity

import (
	"bytes"
	"fmt"
)

type TypeTree struct {
	Version    int32
	Depth      int8
	IsArray    bool
	TypeOffset int32
	Type       string
	NameOffset int32
	Name       string
	Size       int32
	Index      int64
	Flags      int32
	Children   []*TypeTree
}

func readBlobTypeTree(reader *Reader, tt *TypeTree, isLittleEndian bool, formatVer uint32) (err error) {
	var nodeData, data []byte
	numNodes, _ := reader.Uint32()
	bufferBytes, _ := reader.Uint32()
	nodeData, _ = reader.Bytes(24 * int64(numNodes))
	data, err = reader.Bytes(int64(bufferBytes))

	parents := []*TypeTree{tt}
	treeReader, err := NewReader(nodeData)
	treeReader.ChangeEndian(isLittleEndian)
	currentDepth := int16(-1)

	for i := uint32(0); i < numNodes; i++ {
		_, _ = treeReader.Int16()
		depth, _ := treeReader.Uint8()
		isArray, _ := treeReader.Int8()
		typeOffset, _ := treeReader.Int32()
		nameOffset, _ := treeReader.Int32()
		sz, _ := treeReader.Int32()
		index, _ := treeReader.Uint32()
		flags, _ := treeReader.Int32()

		node := &TypeTree{
			Type:    strFromBuf(bufferBytes, data, typeOffset),
			Name:    strFromBuf(bufferBytes, data, nameOffset),
			IsArray: isArray > 0,
			Size:    sz,
			Index:   int64(index),
			Flags:   flags,
		}

		if int16(depth) > currentDepth {
			parents = append(parents, node)
			currentDepth = int16(depth)
			continue
		}

		for i := int16(0); i < currentDepth-int16(depth)+1; i++ {
			lastIdx := len(parents) - 1
			last := parents[lastIdx]
			parents = parents[:lastIdx]
			parents[lastIdx-1].Children = append(parents[lastIdx-1].Children, last)
		}

		if flags&0x4000 > 0 {
			if _, err = treeReader.Align(); err != nil {
				return err
			}
		}

		parents = append(parents, node)
		currentDepth = int16(depth)
	}

	fmt.Println(parents)

	return nil
}

func strFromBuf(sz uint32, buf []byte, offset int32) (str string) {
	var data []byte
	if offset < 0 {
		offset &= 0x7fffffff
		data = STRINGS_DAT
	} else if uint32(offset) < sz {
		data = buf
	} else {
		return ""
	}
	n := int32(bytes.IndexByte(data[offset:], 0))
	return string(data[offset : offset+n])
}

func readOldTypeTree(reader *Reader, tt *TypeTree, isLittleEndian bool) (err error) {
	tt.Type, _ = reader.StringNull()
	tt.Name, _ = reader.StringNull()
	tt.Size, _ = reader.Int32()
	index, _ := reader.Int32()
	isArray, _ := reader.Int32()
	tt.Version, _ = reader.Int32()
	tt.Flags, _ = reader.Int32()
	numChildren, _ := reader.Int32()

	tt.Index = int64(index)
	tt.IsArray = isArray > 0

	for i := int32(0); i < numChildren; i++ {
		node := &TypeTree{}
		if err = readOldTypeTree(reader, node, isLittleEndian); err != nil {
			return err
		}
		tt.Children = append(tt.Children, node)
	}

	return nil
}

func ReadTypeTree(reader *Reader, isLittleEndian bool, formatVer uint32) (*TypeTree, error) {
	tt := &TypeTree{}
	if formatVer == 10 || formatVer >= 12 {
		return tt, readBlobTypeTree(reader, tt, isLittleEndian, formatVer)
	}

	return tt, readOldTypeTree(reader, tt, isLittleEndian)
}
