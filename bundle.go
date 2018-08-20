package unity

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudflare/golz4"
)

const (
	SignatureRaw     = "UnityRaw"
	SignatureWeb     = "UnityWeb"
	SignatureFS      = "UnityFS"
	SignatureArchive = "UnityArchive"
)

const (
	CompressionNone = iota
	CompressionLZMA
	CompressionLZ4
	CompressionLZ4HC
	CompressionLZHAM
	CompressionMax
)

type Bundle struct {
	Signature        string
	FormatVersion    int32
	TargetVersion    string
	GeneratorVersion string

	// Raw
	FileSize               uint32
	HeaderSize             uint32
	FileCount              uint32
	BundleCount            uint32
	BundleSize             uint32
	UncompressedBundleSize uint32
	CompressedFileSize     uint32
	AssetHeaderSize        uint32
	NumAssets              uint32

	// FS
	FSFileSize  int64
	CIBlockSize uint32
	UIBlockSize uint32

	Flags           uint32
	CompressionType int
	Name            string
	Assets          []*Asset
}

func ReadBundle(path string) (*Bundle, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader, err := NewReader(buf)
	if err != nil {
		return nil, err
	}

	bundle := &Bundle{}
	if bundle.Signature, err = reader.StringNull(); err != nil {
		return nil, err
	}
	if bundle.FormatVersion, err = reader.Int32(); err != nil {
		return nil, err
	}
	if bundle.TargetVersion, err = reader.StringNull(); err != nil {
		return nil, err
	}
	if bundle.GeneratorVersion, err = reader.StringNull(); err != nil {
		return nil, err
	}

	switch bundle.Signature {
	case SignatureRaw:
	case SignatureWeb:
		return bundle, readRaw(bundle, reader)
	case SignatureFS:
		return bundle, readFS(bundle, reader)
	case SignatureArchive:
		return bundle, readArchive(bundle, reader)
	}

	return nil, fmt.Errorf("Unsupported signature: %s", bundle.Signature)
}

func (bundle *Bundle) ResolveAsset(index int) error {
	return bundle.Assets[index].LoadObjects(bundle.Signature)
}

func (bundle *Bundle) Compressed() bool {
	return bundle.Signature == SignatureWeb
}

func (bundle *Bundle) Decompress(reader *Reader, compressionType int) ([]byte, error) {
	data, _ := reader.Bytes(int64(bundle.CIBlockSize))
	switch compressionType {
	case CompressionNone:
		return data, nil
	case CompressionLZ4:
	case CompressionLZ4HC:
		decompressed := make([]byte, int(bundle.UIBlockSize))
		return decompressed, lz4.Uncompress(data, decompressed)
	}

	return nil, fmt.Errorf("unity.Bundle.Decompress: Unsupported compression type: %v", compressionType)
}

func readRaw(bundle *Bundle, reader *Reader) (err error) {
	bundle.FileSize, _ = reader.Uint32()
	bundle.HeaderSize, _ = reader.Uint32()
	bundle.FileCount, _ = reader.Uint32()
	bundle.BundleCount, _ = reader.Uint32()

	if bundle.FormatVersion >= 2 {
		bundle.BundleSize, _ = reader.Uint32()
		if bundle.FormatVersion >= 3 {
			bundle.UncompressedBundleSize, _ = reader.Uint32()
		}
	}

	if bundle.HeaderSize >= 60 {
		bundle.CompressedFileSize, _ = reader.Uint32()
		bundle.AssetHeaderSize, _ = reader.Uint32()
	}

	_, _ = reader.Int32()
	_, _ = reader.Int8()

	bundle.Name, _ = reader.StringNull()

	reader.SeekStart(int64(bundle.HeaderSize))

	numAssets := uint32(1)
	if bundle.Compressed() {
		numAssets, _ = reader.Uint32()
	}

	for i := uint32(0); i < numAssets; i++ {
		bundle.Assets = append(bundle.Assets, AssetFromBundle(bundle, reader))
	}

	return nil
}

func readFS(bundle *Bundle, reader *Reader) (err error) {
	bundle.FSFileSize, _ = reader.Int64()
	bundle.CIBlockSize, _ = reader.Uint32()
	bundle.UIBlockSize, _ = reader.Uint32()

	flags, _ := reader.Uint32()
	compressionType := int(flags & 0x3F)
	eofMetadata := (flags & 0x80) > 0
	if eofMetadata {
		fmt.Println("eof")
	}

	bundleData, err := bundle.Decompress(reader, compressionType)
	bundleReader, err := NewReader(bundleData)
	if err != nil {
		return err
	}

	_, _ = bundleReader.Bytes(16)

	numBlocks, _ := bundleReader.Int32()
	blocks := []ArchiveBlockInfo{}
	for i := 0; i < int(numBlocks); i++ {
		buSize, _ := bundleReader.Int32()
		bcSize, _ := bundleReader.Int32()
		flags, _ := bundleReader.Int16()
		blocks = append(blocks, ArchiveBlockInfo{buSize, bcSize, flags})
	}

	numNodes, _ := bundleReader.Int32()
	nodes := []ArchiveNode{}
	for i := 0; i < int(numNodes); i++ {
		offset, _ := bundleReader.Int64()
		size, _ := bundleReader.Int64()
		status, _ := bundleReader.Int32()
		name, _ := bundleReader.StringNull()
		nodes = append(nodes, ArchiveNode{offset, size, status, name})
	}

	for i, node := range nodes {
		reader.SeekCurrent(node.Offset)
		block := blocks[i]
		compressed, _ := reader.Bytes(int64(block.CompressedSize))
		decompressed, err := block.Decompress(compressed)
		if err != nil {
			panic(err)
		}
		currentBlock, _ := NewReader(decompressed)
		asset := AssetFromBundle(bundle, currentBlock)
		asset.Name = node.Name
		bundle.Assets = append(bundle.Assets, asset)
	}

	bundle.Name = nodes[0].Name

	return nil
}

func readArchive(bundle *Bundle, reader *Reader) (err error) {
	return errors.New("UnityArchive currently not supported.")
}
