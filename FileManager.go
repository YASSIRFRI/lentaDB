package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)


type FileHeader interface {
	WriteHeader(io.Writer) error
	ReadHeader(io.ReaderAt) error
	Size() int64
}

type FileManager struct{
	directory string
	fileheader FileHeader
	WritePointer *os.File
	logPointer *os.File
	ReadPointer int64
	MaxFileSize int64
}


func (fl *FileManager) init() error {
	if fl.logPointer== nil{
		log:=filepath.Join(fl.directory, "log")
		file, err := os.OpenFile(log, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			return errors.New("Error opening log file")
		}
		fl.logPointer=file
	}
	c:=make(chan bool)
	go fl.flushLog(c)
	test:=<-c
	if test == false {
		return errors.New("Error flushing log file")
	}
	directoryContent, err := ioutil.ReadDir(fl.directory)
	if err != nil {
		return errors.New("Error reading directory")
	}
	files:=make([]os.FileInfo, 0)
	for _, file := range directoryContent {
		if file.Name() == "log" {
			continue
		}
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})
	files = files[1:]
	for _, file := range files {
		filePath := filepath.Join(fl.directory, file.Name())
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0755)
		if err != nil {
			return errors.New("Error opening file for writing")
		}
		err = fl.ValidateFile(file)
		if err != nil {
			fmt.Println("Corrupted File Detected, cannot recover :(")
		}
	}
	return nil
}

func (fl *FileManager) ValidateFile(file *os.File) error {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	if fileSize < 16 {
		return errors.New("File is too small to validate")
	}
	buffer := make([]byte, fileSize-16)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}
	hash := md5.New()
	_, err = hash.Write(buffer)
	if err != nil {
		return err
	}
	calculatedHash := hex.EncodeToString(hash.Sum(nil))
	expectedHashBytes := make([]byte, 16)
	_, err = file.ReadAt(expectedHashBytes, fileSize-16)
	if err != nil {
		return err
	}
	hashString := hex.EncodeToString(expectedHashBytes)
	if calculatedHash != hashString {
		fmt.Println(file.Name())
		fmt.Println(calculatedHash, hashString)
		return errors.New("File validation failed")
	}
	return nil
}

func NewFileManager() (*FileManager, error) {
	directory := "data"
	f := FileManager{directory: directory}
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, errors.New("Directory does not exist")
	}
	d, err := os.Open(directory)
	if err != nil {
		return nil, errors.New("Error opening directory")
	}
	dircontent, err := d.Readdir(-1)
	files:=make([]os.FileInfo, 0)
	for _, file := range dircontent {
		if file.Name() == "log" {
			continue
		}
		files = append(files, file)
	}
	if err != nil {
		return nil, errors.New("Error reading directory")
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})
	var lastFile os.FileInfo
	if len(files)>0{
		lastFile = files[0]
		filePath := filepath.Join(directory, lastFile.Name())
		writePtr, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0755)
		if err != nil {
			return nil, errors.New("Error opening file for writing")
		}
		f.WritePointer = writePtr
	}else{
		newfile ,err := f.createNewFile()
		if err != nil {
			return nil, err
		}
		f.WritePointer = newfile
	}
	if err != nil {
		return nil, errors.New("Error reading file info")
	}
	log:=filepath.Join(f.directory, "log")
	file, err := os.OpenFile(log, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil,errors.New("Error opening log file")
	}
	f.logPointer=file
	if err != nil {
		return nil, errors.New("Error initializing file manager")
	}

	return &f, nil
}

func (f *FileManager) Write(data []byte) error {
	if f.WritePointer == nil {
		fmt.Println("Write pointer is nil")
		newfile, err := f.createNewFile()
		if err != nil {
			fmt.Println("Error creating new file")
			return err
		}
		f.WritePointer = newfile
	}
	_, err := f.WritePointer.Write(data)
	return err
}


func (f *FileManager) createNewFile() (*os.File,error) {
	filePath := filepath.Join(f.directory, strconv.FormatInt(time.Now().UnixNano(), 10))
	filePath = filePath + ".sst"
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating new file 1")
		return nil,err
	}
	if err := f.writeHeader(file); err != nil {
		fmt.Println("Error writing header")
		return nil,err
	}
	return file,nil
}

func (f *FileManager) writeHeader(file *os.File) error {
	header:=NewSSTHeader()
	header.Timestamp=time.Now()
	header.Version=1
	header.size=50
	err:=header.WriteHeader(file);
	return err
}



func (f *FileManager) Log(logRecord []byte) (error){
	directory := f.directory
	fmt.Println("Log file pointer", f.logPointer)
	if f.logPointer == nil {
		log:=filepath.Join(directory, "log")
		file, err := os.OpenFile(log, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			return errors.New("Error opening log file")
		}
		f.logPointer=file
	}
	_ , err:=f.logPointer.Write(logRecord)
	if err != nil {
		fmt.Println("Error writing to log file")
		return err
	}
	return nil
}

func (f *FileManager)Read()(map[string]Entry, error) {
    header := f.fileheader
    err := header.ReadHeader(f.WritePointer)
	fmt.Println("f.WritePointer", f.WritePointer.Name())
    if err != nil {
        fmt.Println("Error reading header")
        return nil, err
    }
    offset := header.Size()
    mp := make(map[string]Entry)
    fileInfo, err := f.WritePointer.Stat()
    if err != nil {
        return nil, err
    }
    for offset < fileInfo.Size(){
		fmt.Println("Offset", offset)
        if offset+2 > fileInfo.Size() {
            break
        }
        entrySizeBytes := make([]byte, 2)
        _, err := f.WritePointer.ReadAt(entrySizeBytes, offset)
        if err != nil {
            fmt.Println("Error reading entry size")
            return nil, err
        }
        entrySize := binary.BigEndian.Uint16(entrySizeBytes)
        if offset+2+int64(entrySize) > fileInfo.Size() {
            break
        }
        entryData := make([]byte, entrySize)
        _, err = f.WritePointer.ReadAt(entryData, offset+2)
        if err != nil {
            fmt.Println("Error reading entry data")
            return nil, err
        }
        entryType := entryData[0]
        entryData = entryData[1:]
        s := string(entryData)
		//fmt.Println("Entry", s)
		//fmt.Println("Entry type", entryType)
        split := strings.Split(s, "=")
        if entryType == 0 {
			mp[split[0]] = Entry{Key:split[0],Value:split[1],t:0}
        }else{
			mp[split[0]] = Entry{Key:split[0],Value:split[1],t:1}
		}
        offset += int64(entrySize) + 2
    }
    return mp, nil
}


func (f *FileManager) loadFile(filetoLoad *os.File) (map[string]Entry, error) {
    header := f.fileheader
    err := header.ReadHeader(filetoLoad)
    if err != nil {
        fmt.Println("Error reading header")
        return nil, err
    }
    offset := header.Size()
    mp := make(map[string]Entry)
    checksumSize := int64(16)
	fileInfo, err := filetoLoad.Stat()
	if err != nil {
		return nil, err
	}
	fmt.Println("Reading from File size", fileInfo.Size())
    for offset < fileInfo.Size()-checksumSize{
        entrySizeBytes := make([]byte, 2)
        _, err := filetoLoad.ReadAt(entrySizeBytes, offset)
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Println("Error reading entry size")
            return nil, err
        }
        entrySize := binary.BigEndian.Uint16(entrySizeBytes)
        entryData := make([]byte, entrySize)
        _, err = filetoLoad.ReadAt(entryData, offset+2)
        if err != nil {
            fmt.Println("Error reading entry data")
            return nil, err
        }
		entryType := entryData[0]
		entryData = entryData[1:]
		s := string(entryData)
		split := strings.Split(s, "=")
		if entryType == 0 {
			mp[split[0]] = Entry{Key:split[0],
				Value:split[1],t:0}
		}else{
			mp[split[0]] = Entry{Key:split[0],Value:split[1],t:1}
		}
        offset += int64(entrySize) +2  
    }
    //checksumBytes := make([]byte, checksumSize)
    //_, err = filetoLoad.ReadAt(checksumBytes, offset)
    //if err != nil {
        //fmt.Println("Error reading checksum")
        //return nil, err
    //}
    return mp, nil
}


func (f *FileManager) flushLog(c chan<- bool) error {
	logInfo, err := f.logPointer.Stat()
	if err != nil {
		fmt.Println("Error reading log file")
		fmt.Println(err)	
		c<-false
		return err
	}
	logsize:=logInfo.Size()
    logcontent := make([]byte, logsize)
    _, err = f.logPointer.ReadAt(logcontent, 0)
    if err != nil {
		fmt.Println("Error read content of the  log file")
		fmt.Println(err)
		c<-false
        return err
    }
    if err != nil {
		fmt.Println("Error obtaining lock for log file")
		c<-false
        return err
    }
	fmt.Println("Log content", len(logcontent))
    err =f.Write(logcontent)
    if err != nil {
		fmt.Println("Error writing to SST file")
		fmt.Println(err)
		c<-false
        return err
    }
	logmutex:=sync.Mutex{}
	logmutex.Lock()
	err=f.logPointer.Close()
	if(err!=nil){
		fmt.Println("Error closing log file")
		fmt.Println(err)
		c<-false
		return err
	}
	err = os.Truncate(f.logPointer.Name(), 0)
	f.logPointer ,err = os.OpenFile(f.logPointer.Name(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Error recreating the  log file")
		fmt.Println(err)
		c<-false
		return err
	}
	logmutex.Unlock()
	if(err!=nil){
		fmt.Println("Error truncating log file")
		fmt.Println(err)
		c<-false
		return err
	}
	c<-true
    return nil
}

func (f *FileManager) closeFile() error{
	fileInfo, err := f.WritePointer.Stat()
	if err != nil {
		return err
	}
	fileContent := make([]byte, fileInfo.Size())
	_,err=f.WritePointer.ReadAt(fileContent, 0)
	if err != nil {
		fmt.Println("Error reading file content")
		return err
	}
	hasher := md5.New()
	hasher.Write(fileContent)
	fmt.Println("Hash", hasher.Sum(nil))
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)
	fmt.Println("Hex MD5 Hash:", hashString)
	logMutes:=sync.Mutex{}
	logMutes.Lock()
	_,err=f.WritePointer.Write(hash)
	if err != nil {
		fmt.Println("Error writing hash")
		return err
	}
	logMutes.Unlock()
	f.WritePointer.Close()
	return nil
}

func (f *FileManager) compact() error {
    // Get a list of SST files in the directory
    files, err := ioutil.ReadDir(f.directory)
    if err != nil {
        return err
    }
    sstFiles := []os.FileInfo{}
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".sst") {
            sstFiles = append(sstFiles, file)
        }
    }
    if len(sstFiles) <= 10 {
        return nil
    }
    globalMap := make(map[string]Entry)
    for _, file := range sstFiles {
        filePath := filepath.Join(f.directory, file.Name())
		fileToCompact, err := os.Open(filePath)
		if err != nil {
			return err
		}
        fileMap, err := f.loadFile(fileToCompact)
        if err != nil {
            return err
        }
        for _, entry := range fileMap {
			if entry.t == 0 {
				globalMap[entry.Key] = entry
			}else{
				delete(globalMap, entry.Key)
			}
        }
    }
    compactedFile, err := f.createNewFile()
    if err != nil {
        return err
    }
    defer compactedFile.Close()
	buffer:=make([]byte, 0)
    for _, entry := range globalMap {
        entryBytes := entry.toBytes()
		buffer=append(buffer, entryBytes...)
    }
    _, err = compactedFile.Write(buffer)
    if err != nil {
        return err
    }
    compactedFile.Close()
    for _, file := range sstFiles {
        err := os.Remove(filepath.Join(f.directory, file.Name()))
        if err != nil {
            return err
        }
    }
    return nil
}

func (f *FileManager) flushMem(mem *MemTable) error {
	buffer:=make([]byte, 0)
	for _, v := range mem.Memdata {
		entry:=v.toBytes()
		buffer=append(buffer, entry...)
	}
	f.Write(buffer)
	f.closeFile()
	//f.compact()	
	newFile,err:=f.createNewFile()
	if err != nil {
		return err
	}
	f.WritePointer=newFile
	err = os.Truncate(f.logPointer.Name(), 0)
	f.logPointer ,err = os.OpenFile(f.logPointer.Name(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return errors.New("Error recreating the log file")
	}
	return nil
}
