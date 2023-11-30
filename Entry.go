package main


import (
	"encoding/binary"
)

type Entry struct {
    Key string
    Value string
    t int 
}

func (e *Entry) toBytes() []byte {
    entrySize := 2+len(e.Key)+len(e.Value)+1+1
    entry := make([]byte, entrySize)
    binary.BigEndian.PutUint16(entry, uint16(len(e.Key)+len(e.Value)+2))
    entry[2] = byte(e.t)
    copy(entry[3:], e.Key)
    entry[len(e.Key)+3] = 61
    copy(entry[len(e.Key)+4:], e.Value)
    return entry
}