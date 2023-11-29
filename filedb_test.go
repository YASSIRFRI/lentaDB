package main


import (
    "testing"
    "strconv"
)

func TestSet(t *testing.T) {
    f1, err := NewFileManager()
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    f1.MaxLogSize = 100
    f1.MaxFileSize = 100
    f1.init()
    db, err := NewFileDB(f1)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    db.MaxEntrySize = 100
    db.CacheSize = 100

    // Test case 1: Entry size too large
    key := []byte("qwerqwerqwerqwerqwerqwerqwreqwerqwreqwerqwerqwerqwerqwerqwerqwerqwerqwer")
    value := []byte("asdfasdfasdfasdfasdfasdfasdfasfdasfasdfasdfasfasdfasdfasdfasfdadsaf")
    err = db.Set(key, value)
    if err == nil {
        t.Error("Expected an error for large entry size, but got nil")
    }

    // Test case 2: Valid entry
    key = []byte("validKey")
    value = []byte("validValue")
    err = db.Set(key, value)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    // Test case 3: Cache overflow
    for i := 0; i < 110; i++ {
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
