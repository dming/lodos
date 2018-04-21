package seqconf

import (
	"bytes"
	"os"
	"bufio"
	"strings"
	"encoding/json"
	"github.com/dming/lodos/addon/seq"
)

var (
	Conf = &Config{}
)

type Config struct {
	Alloc Alloc
	Store Store
	Mediate Mediate
	Settings map[string]interface{}
}

type Alloc struct {
	Host Host
}

type Mediate struct {
	Host Host
}

type Store struct {
	Host Host
	Dir  string
}

type Host struct {
	IP string
	Port int
}

// json配置格式的路由表
/*
{
  "version":1,
  "sets": [
    "set_id":1,
    "set_name":"0-100",
    "range": {
      "id":0,
      "size":10
    },
    "allocs": [
      "alloc_id":,
      "addr":,
      "alloc_name":,
      "ranges":[
      ],
    ]
  ]
}
*/
type Router struct {
	Version int
	Sets []*Set
	Allocs []*Alloc
	Mediates []*Mediate
	Stores []*Store
}

type MaxSeq struct {
	Version int
	Max map[string]seq.Sequence
}


type Set struct {
	Set_id int
	Set_name string
	Range
}

type Range struct {
	ID int
	Size int
}


func LoadConfig(Path string) {
	// Read config.
	if err := readFileInto(Path); err != nil {
		panic(err)
	}
	//
}

func readFileInto(path string) error {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//log.Info(string(data))
	return json.Unmarshal(data, &Conf)
}