package storage

import "encoding/binary"

func WriteInt64(data int64) (value []byte) {
	value = make([]byte, 8)
	binary.BigEndian.PutUint64(value, uint64(data))
	return
}

func ReadInt64(data []byte) int64 {
	return int64(binary.BigEndian.Uint64(data))
}
