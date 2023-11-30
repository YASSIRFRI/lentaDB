package main

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestSet(t *testing.T) {
	t.Run("LargeEntrySizeError", testLargeEntrySizeError)
	t.Run("ValidEntry", testValidEntry)
	t.Run("CacheOverflow", testCacheOverflow)
	t.Run("WriteAndVerifyFiles", testWriteAndVerifyFiles)
}

func testLargeEntrySizeError(t *testing.T) {
	f1, err := NewFileManager()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	db, err := NewFileDB(f1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	db.MaxEntrySize = 100
	db.CacheSize = 100
	key := []byte("qwerqwerqwerqwerqwerqwerqwreqwerqwreqwerqwerqwerqwerqwerqwerqwerqwerqwerqwerqwerqwerqwerqwer")
	value := []byte("asdfasdfasdfasdfasdfasdfasdfasfdasfasdfasdfasfasdfasdfasdfasfdadsafasdfasdfasdfasdfasdfasdfasdf")
	err = db.Set(key, value)
	if err == nil {
		t.Error("Expected an error for large entry size, but got nil")
	}
}

func testValidEntry(t *testing.T) {
	f1, err := NewFileManager()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	db, err := NewFileDB(f1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	db.MaxEntrySize = 100
	db.CacheSize = 100
	key := []byte("validKey")
	value := []byte("validValue")
	err = db.Set(key, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func testCacheOverflow(t *testing.T) {
	f1, err := NewFileManager()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test case 3: Cache overflow
	db, err := NewFileDB(f1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	db.MaxEntrySize = 100
	db.CacheSize = 100
	for i := 0; i < 100; i++ {
		key := []byte("key" + strconv.Itoa(i))
		value := []byte("value" + strconv.Itoa(i))
		err := db.Set(key, value)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	// Check that the cache size is maintained
	if len(db.MemTable.Memdata) != db.CacheSize {
		t.Errorf("Expected cache size %d, got %d", db.CacheSize, len(db.MemTable.Memdata))
	}
}

func testWriteAndVerifyFiles(t *testing.T) {
	buffer := bytes.Buffer{}
	sstHeader := NewSSTHeader()
	err := sstHeader.WriteHeader(&buffer)
	if err != nil {
		t.Errorf("Unexpected error writing SST header: %v", err)
	}
	bufferBytes, err := ioutil.ReadAll(&buffer)
	if err != nil {
		t.Errorf("Unexpected error converting buffer to bytes.Reader: %v", err)
	}
	bufferReader := bytes.NewReader(bufferBytes)
	err = sstHeader.ReadHeader(bufferReader)
	if err != nil {
		t.Errorf("Unexpected error reading SST header: %v", err)
	}
}
