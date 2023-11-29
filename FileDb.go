package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)



type FileDB struct {
    FileManager *FileManager
    MemTable *MemTable
    MaxEntrySize int
    CacheSize int
}

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



func (fl *FileDB) exists(key []byte) ([]byte, error) {
    if v, ok := fl.MemTable.Memdata[string(key)]; ok {
        if v.t == 1 {
            return nil, nil
        } else {
            return []byte(v.Value), nil
        }
    }
    it := DirectoryIterator{file: fl.FileManager.WritePointer}
    mp, err := fl.FileManager.Read()
    if err != nil {
        fmt.Println("Error in exists 4")
        return nil, err
    }
    fmt.Println(mp)
    if v, ok := mp[string(key)]; ok {
        if v.t == 1 {
            return nil, nil
        } else {
            return []byte(v.Value), nil
        }
    }

    for {
        err := it.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Println("Error in exists 3")
            return nil, err
        }

        mp, err := fl.FileManager.loadFile(it.file)
        if err != nil {
            fmt.Println("Error in exists 2")
            return nil, err
        }
        fmt.Println(mp)
        fmt.Println(string(key))
        v, ok := mp[string(key)]
        fmt.Println(v, ok)
        if v, ok := mp[string(key)]; ok {
            if v.t == 1 {
                return nil, nil
            } else {
                return []byte(v.Value), nil
            }
        }
        if err != nil {
            fmt.Println("Error in exists 1")
            return nil, err
        }
    }

    return nil, nil
}

func (fl *FileDB) Set(key, value []byte) error {
    if len(key) > fl.MaxEntrySize || len(value) > fl.MaxEntrySize {
        return errors.New("Entry size too large")
    }
    fl.MemTable.Memdata[string(key)] = Entry{Key: string(key), Value: string(value), t: 0}
    //2 for the size 
    //1 for the type of the entry
    //1 for the = sign
    logSize:=2+len(key)+len(value)+1+1;
    logEntry:=make([]byte, logSize)
    binary.BigEndian.PutUint16(logEntry, uint16(len(key)+len(value)+2))
    logEntry[2]=0
    copy(logEntry[3:], key)
    logEntry[len(key)+3]=61
    copy(logEntry[len(key)+4:], value)
    err:=fl.FileManager.Log(logEntry);
    if err != nil {
        return err
    }
    //fmt.Println(logEntry)
    //fmt.Println(unsafe.Sizeof(fl.MemTable.Memdata)+unsafe.Sizeof(fl.MemTable.DeletedItems))
    if len(fl.MemTable.Memdata)> fl.CacheSize {
        //fl.FileManager.flushLog()
        //fl.MemTable.Memdata = make(map[string][]byte)
        fl.FileManager.flushMem(fl.MemTable);
    }
    return  nil
}
func (fl *FileDB) Get(key []byte) ([]byte, error) {
    v, err := fl.exists(key)
    //fmt.Println(v)
    if err != nil {
        return nil, err
    }
    if v != nil {
        return v, nil
    }
    fmt.Println("Key Not found!")
    return nil, nil
}


func (fl *FileDB) Del(key []byte) ([]byte, error) {
    if v, err:=fl.exists(key); err != nil {
        return nil, err
    } else if v != nil {
        fl.MemTable.Memdata[string(key)]=Entry{Key: string(key), Value: string(v), t: 1}
        logSize:=2+len(key)+len(v)+1+1;
        logEntry:=make([]byte, logSize)
        binary.BigEndian.PutUint16(logEntry, uint16(len(key)+len(v)+2))
        logEntry[2]=1
        copy(logEntry[3:], key)
        logEntry[len(key)+3]=61
        copy(logEntry[len(key)+4:], v)
        err:=fl.FileManager.Log(logEntry);
        if err != nil {
            return nil, err
        }
        return v, nil
    }
    return nil, nil
}

func NewFileDB(f *FileManager) (*FileDB, error) {
    MemTable:=&MemTable{
        Memdata: make(map[string]Entry),
    }
    return &FileDB{
        FileManager: f,
        MemTable: MemTable,
    }, nil
}

