package ut

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"strconv"
	"strings"
)

func Hash(p []byte) []byte {
	v := make([]byte, 4)
	hash := fnv.New32a()
	hash.Write(p)
	binary.BigEndian.PutUint32(v, hash.Sum32())
	return v
}

func ClearBit(b byte, pos uint) byte {
	return b &^ (1 << pos)
}

func SetBit(b byte, pos uint) byte {
	return b | ^(1 << pos)
}

func FixBoundary(listLen, index int64) int64 {
	if listLen == 0 {
		return 0
	}
	if index < -1*listLen {
		index = 0
	} else if index < 0 {
		index += listLen
	} else if index >= listLen {
		index = listLen - 1
	}
	return index
}

func LowerString(s []byte) string {
	return strings.ToLower(string(s))
}

func ParseInt64(v []byte) (int64, error) {
	return strconv.ParseInt(string(v), 10, 64)
}

func ParseFloat64(v []byte) (float64, error) {
	return strconv.ParseFloat(string(v), 64)
}

func FormatInt64(v int64) []byte {
	return []byte(strconv.FormatInt(v, 10))
}

func Float642Byte(v float64) []byte {
	return []byte(strconv.FormatFloat(v, 'f', -1, 64))
}

func ConcatBytearray(bytesarray ...[]byte) []byte {
	var totalLen int
	for _, s := range bytesarray {
		totalLen += len(s)
	}
	buf := make([]byte, totalLen)

	var i int
	for _, s := range bytesarray {
		i += copy(buf[i:], s)
	}
	return buf
}

func Bytes2Int64(val []byte) (int64, error) {
	var i int64
	buf := bytes.NewBuffer(val)
	err := binary.Read(buf, binary.LittleEndian, &i)
	return i, err
}

func Int642Bytes(val int64) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &val)
	return buf.Bytes()
}

func DecodeInt64(val []byte) ([]byte, error) {
	i, err := Bytes2Int64(val)
	if err != nil {
		return nil, err
	}
	return FormatInt64(i), nil
}
