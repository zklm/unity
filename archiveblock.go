package unity

import (
	"fmt"

	lz4 "github.com/cloudflare/golz4"
)

type ArchiveNode struct {
	Offset int64
	Size   int64
	Status int32
	Name   string
}

type ArchiveBlockInfo struct {
	UncompressedSize int32
	CompressedSize   int32
	Flags            int16
}

func (block *ArchiveBlockInfo) Decompress(data []byte) ([]byte, error) {
	comp := int(block.Flags & 0x3F)
	switch comp {
	case CompressionNone:
		return data, nil
	case CompressionLZMA:
		return DecompressLZMARaw(data, block.UncompressedSize)
	case CompressionLZ4:
	case CompressionLZ4HC:
		decompressed := make([]byte, int(block.UncompressedSize))
		return decompressed, lz4.Uncompress(data, decompressed)
	}
	return nil, fmt.Errorf("unity.ArchiveBlockInfo.Decompress: Unsupported compression type: %v", comp)
}

type ArchiveBlockStorage struct {
	*Reader
	Blocks             []ArchiveBlockInfo
	VirtualSize        int64
	Cursor             int64
	BaseOffset         int64
	CurrentBlockIndex  int
	CurrentBlockOffset int64
	CurrentBlock       *Reader
}

func NewArchiveBlockStorage(reader *Reader, blocks []ArchiveBlockInfo) *ArchiveBlockStorage {
	storage := ArchiveBlockStorage{
		Reader:            reader,
		Blocks:            blocks,
		BaseOffset:        reader.Tell(),
		CurrentBlockIndex: -1,
	}

	for _, block := range blocks {
		storage.VirtualSize += int64(block.UncompressedSize)
	}

	return &storage
}

func (storage *ArchiveBlockStorage) inCurrentBlock(pos int64) bool {
	if storage.CurrentBlockIndex < 0 {
		return false
	}
	off := storage.CurrentBlockOffset
	end := off + int64(storage.Blocks[storage.CurrentBlockIndex].UncompressedSize)
	return off <= pos && pos < end
}

func (storage *ArchiveBlockStorage) seekToBlock(pos int64) {
	if storage.inCurrentBlock(pos) {
		return
	}

	baseOffset := int64(0)
	offset := int64(0)
	found := false
	for i, b := range storage.Blocks {
		if offset+int64(b.UncompressedSize) > pos {
			storage.CurrentBlockIndex = i
			found = true
			break
		}
		baseOffset += int64(b.CompressedSize)
		offset += int64(b.UncompressedSize)
	}

	if !found {
		storage.CurrentBlockIndex = -1
		storage.CurrentBlock.Reset()
		storage.CurrentBlock = nil
		return
	}

	storage.CurrentBlockOffset = offset
	storage.Reader.SeekStart(storage.BaseOffset + baseOffset)

	currentBlock := storage.Blocks[storage.CurrentBlockIndex]
	compressed, _ := storage.Reader.Bytes(int64(currentBlock.CompressedSize))
	if currentBlockB, err := currentBlock.Decompress(compressed); err != nil {
		fmt.Println(err)
	} else {
		storage.CurrentBlock, _ = NewReader(currentBlockB)
	}
}
