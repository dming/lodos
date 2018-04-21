package base

import (
	"github.com/dming/lodos/addon/seq"
	"github.com/dming/lodos/addon/seq/seqconf"
	"os"
	"fmt"
	"bytes"
	"bufio"
	"strings"
	"encoding/json"
	"io/ioutil"
)

//主要为将alloc_svr修改的数据写到redis数据库中
//系统重启的时候也需要从数据库中读取数据到alloc里

func NewStoreSvr(settings interface{}) seq.StoreSvr {
	return &storeSvr{}
}

type storeSvr struct {
	settings *seqconf.Store

	max_seqs *seqconf.MaxSeq
	maxseq_err_chan chan bool
}

func (s *storeSvr) OnInit(settings *seqconf.Store) {
	//LoadFiles(settings.Dir)
	s.settings = settings
}

func (s *storeSvr) Run() {
	panic("implement me")
}

func (s *storeSvr) OnDestroy() {
	panic("implement me")
}

func (s *storeSvr) LoadRouterTable() (seq.RouterTable, error) {
	panic("implement me")
	//todo: load table file, format from json to struct
	osdir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	routerPath := fmt.Sprintf("%s/%s/router.dat", osdir, s.settings.Dir)
	file, err := os.Open(routerPath)
	if err != nil {
		return nil, fmt.Errorf("cannot find the router file")
	}
	defer file.Close()

	var data []byte
	buf := new(bytes.Buffer)
	r := bufio.NewReader(file)
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
	table := &routerTable{}
	err = json.Unmarshal(data, table)
	if err != nil {
		return nil, err
	}

	return table, nil
}

func (s *storeSvr) SaveRouterTable(table seq.RouterTable) error {
	panic("implement me")
	osdir, err := os.Getwd()
	if err != nil {
		return err
	}
	routerPath := fmt.Sprintf("%s/%s/router.dat", osdir, s.settings.Dir)


	body, err := json.Marshal(table)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(routerPath, body, os.ModeExclusive)
	return err
}

func (s *storeSvr) GetMaxSeq(id int, size int) (int, seq.Sequence, error) {
	name := fmt.Sprintf("%d_%d", id, size)
	if re, ok := s.max_seqs.Max[name]; ok {
		return s.max_seqs.Version, re, nil
	}
	return 0, -1, fmt.Errorf("cannot find %s", name)
}

func (s *storeSvr) LoadMaxSeq() error {
	panic("implement me")
	osdir, err := os.Getwd()
	if err != nil {
		return err
	}
	routerPath := fmt.Sprintf("%s/%s/maxseq.db", osdir, s.settings.Dir)
	file, err := os.Open(routerPath)
	if err != nil {
		return fmt.Errorf("cannot find the router file")
	}
	defer file.Close()

	var data []byte
	buf := new(bytes.Buffer)
	r := bufio.NewReader(file)
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
	var m map[string]seq.Sequence
	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	return nil
}

//如果失败，需要回退？
func (s *storeSvr) SaveMaxSeq(id int, size int, maxSeq seq.Sequence) error {
	name := fmt.Sprintf("%d_%d", id, size)
	s.max_seqs.Version++
	s.max_seqs.Max[name] = maxSeq

	osdir, err := os.Getwd()
	if err != nil {
		return err
	}
	routerPath := fmt.Sprintf("%s/%s/maxseq.db", osdir, s.settings.Dir)

	body, err := json.Marshal(s.max_seqs)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(routerPath, body, os.ModeExclusive)
	return err
}

func (s *storeSvr) loop_handle(err_chan chan bool) {
	for {
		switch {
		case <-err_chan:
			s.LoadMaxSeq()
		}
	}
}