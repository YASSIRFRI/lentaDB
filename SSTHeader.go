package main

import (
	"fmt"
	"io"
	"time"
)

/*
SSTHeader is the type of the header of the SST file
It could be extended to include more information, and can also be changed to a different format
it should only implement the ReadHeader and WriteHeader methods
*/

type SSTHeader struct {
	size 	int64  //header size
	magic 	[]byte //8 bytes
	Version   int64 //2 bytes
	Timestamp time.Time 
}
func TimeStampToBytes(t time.Time) []byte {
	s := t.UTC().Format(time.RFC3339)
	return []byte(s)
}

// 16 bytes hash
func (s *SSTHeader) WriteHeader(w io.Writer) error {
	header := make([]byte, s.size)
	copy(header[0:], s.magic)
	copy(header[8:], []byte{0, 0, 0, 0, 0, 0, 0, 0})
	copy(header[16:], []byte{0, 0})
	copy(header[18:], TimeStampToBytes(s.Timestamp))
	_, err := w.Write(header)
	if err != nil {
		fmt.Println("Error writing header")
		fmt.Println(err)
	}
	return err
}


func (s *SSTHeader) Size() int64 {
	return s.size
}
func (s *SSTHeader) ReadHeader(r io.ReaderAt) error {
	header := make([]byte, s.size)
	_, err := r.ReadAt(header, 0)
	if err != nil {
		return err
	}
	s.magic = header[0:8]
	s.Version = int64(header[8])
	s.Timestamp, err = time.Parse(time.RFC3339, string(header[18:38]))
	if err != nil {
		fmt.Println("Error parsing timestamp")
		return err
	}
	return nil
}

func NewSSTHeader() *SSTHeader {
	return &SSTHeader{size: 50}
}
