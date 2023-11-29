package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Cmd int

const (
	Get Cmd = iota
	Set
	Del
	Ext
	Unk
)

type Error int

func (e Error) Error() string {
	return "Empty command"
}

const (
	Empty Error = iota
)

type DB interface {
	Set(key []byte, value []byte) error

	Get(key []byte) ([]byte, error)

	Del(key []byte) ([]byte, error)
}



type memDB struct {
	values map[string][]byte
}

func (mem *memDB) Set(key, value []byte) error {
	mem.values[string(key)] = value
	return nil
}

func (mem *memDB) Get(key []byte) ([]byte, error) {
	if v, ok := mem.values[string(key)]; ok {
		return v, nil
	}
	return nil, errors.New("Key not found")
}

func (mem *memDB) Del(key []byte) ([]byte, error) {
	if v, ok := mem.values[string(key)]; ok {
		delete(mem.values, string(key))
		return v, nil
	}
	return nil, errors.New("Key doesn't exist")
}

func NewInMem() *memDB {
	values := make(map[string][]byte)
	return &memDB{
		values,
	}
}

type Repl struct {
	db DB

	in  io.Reader
	out io.Writer
}

func (re *Repl) parseCmd(buf []byte) (Cmd, []string, error) {
	line := string(buf)
	elements := strings.Fields(line)
	if len(elements) < 1 {
		return Unk, nil, Empty
	}
	switch elements[0] {
	case "get":
		return Get, elements[1:], nil
	case "set":
		return Set, elements[1:], nil
	case "del":
		return Del, elements[1:], nil
	case "exit":
		return Ext, nil, nil
	default:
		return Unk, nil, nil
	}
}

func (re *Repl) Start() {
	scanner := bufio.NewScanner(re.in)
	for {
		fmt.Fprint(re.out, "> ")
		if !scanner.Scan() {
			break
		}
		buf := scanner.Bytes()
		cmd, elements, err := re.parseCmd(buf)
		if err != nil {
			fmt.Fprintf(re.out, "%s\n", err.Error())
			continue
		}
		switch cmd {
		case Get:
			if len(elements) != 1 {
				fmt.Fprintf(re.out, "Expected 1 arguments, received: %d\n", len(elements))
				continue
			}
			v, err := re.db.Get([]byte(elements[0]))
			if err != nil {
				fmt.Fprintln(re.out, err.Error())
				continue
			}
			fmt.Fprintln(re.out, string(v))
		case Set:
			if len(elements) != 2 {
				fmt.Printf("Expected 2 arguments, received: %d\n", len(elements))
				continue
			}
			err := re.db.Set([]byte(elements[0]), []byte(elements[1]))
			if err != nil {
				fmt.Fprintln(re.out, err.Error())
				continue
			}
		case Del:
			if len(elements) != 1 {
				fmt.Printf("Expected 1 arguments, received: %d\n", len(elements))
				continue
			}
			v, err := re.db.Del([]byte(elements[0]))
			if err != nil {
				fmt.Fprintln(re.out, err.Error())
				continue
			}
			fmt.Fprintln(re.out, string(v))
		case Ext:
			fmt.Fprintln(re.out, "Bye!")
			return
		case Unk:
			fmt.Fprintln(re.out, "Unkown command")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(re.out, err.Error())
	} else {
		fmt.Fprintln(re.out, "Bye!")
	}
}


/*
	Program Entry point Main:
	interface DB: needs Reader,Writer
	FileManager: read write 
	once instantiated : the FileManager look into the file directory and loads the last file 

*/


func (db *FileDB) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is missing", http.StatusBadRequest)
		return
	}
	value,err:= db.Get([]byte(key))
	if err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	if value == nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "GET result for key %s: %s", key, value)
}

func (db *FileDB) HandleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.FormValue("key")
	value := r.FormValue("value")

	if key == "" || value == "" {
		http.Error(w, "Key or value parameter is missing", http.StatusBadRequest)
		return
	}
	err := db.Set([]byte(key), []byte(value))
	if err != nil {
		http.Error(w, "Error setting key", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "SET success for key %s", key)
}

func (db *FileDB) HandleDel(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is missing", http.StatusBadRequest)
		return
	}
	_,err:=db.Del([]byte(key))
	if err != nil {
		http.Error(w, "Error deleting key", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "DEL success for key %s", key)
}


func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	maxFileSize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	FileManager, err := NewFileManager()
	FileManager.MaxFileSize = int64(maxFileSize)
	err = FileManager.init()
	if err != nil {
		fmt.Println(err)
		return
	}
	header := NewSSTHeader()
	FileManager.fileheader = header
	db, err := NewFileDB(FileManager)
	maxEntrySize, _ := strconv.Atoi(os.Getenv("MAX_ENTRY_SIZE"))
	cacheSize, _ := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	fmt.Println("Cache Size: ", cacheSize)
	db.MaxEntrySize = maxEntrySize
	db.CacheSize = cacheSize
	if err != nil {
		fmt.Println(err)
		return
	}
	FileManager.fileheader=header;
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(FileManager)
	if err != nil {
		fmt.Println(err)
		return	
	}
	//repl := &Repl{
		//db:  db,
		//in:  os.Stdin,
		//out: os.Stdout,
	//}
	//repl.Start()
	http.HandleFunc("/get", db.HandleGet)
	http.HandleFunc("/set", db.HandleSet)
	http.HandleFunc("/del", db.HandleDel)
	port := 8080
	fmt.Printf("Server started on :%d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}