package base

import (
	"fmt"
	"github.com/dming/lodos/addon/seq"
)

func NewSection(_flag int, _maxSeq seq.Sequence, _step seq.Sequence) (*section, error) {
	if _step <= 0 || _maxSeq <= 0 {
		return nil, fmt.Errorf("step or max_seq cannot equal or negative than 0")
	}
	return &section{
		begin_id: _flag,
		users:    make(map[int]*user),
		maxSeq:   _maxSeq,
		step:     _step,
	}, nil
}

//user id is between begin_id and begin_id + 1000
type section struct {
	begin_id int
	users    map[int]*user
	maxSeq   seq.Sequence
	step     seq.Sequence
}

func (s *section) StartId() int {
	return s.begin_id
}

func (s *section) AddUser(id int, _seq seq.Sequence) error {
	if _, ok := s.users[id]; ok {
		return fmt.Errorf("user[%s] already exist")
	}
	s.users[id] = &user {
		uid:    id,
		curSeq: _seq,
	}
	return nil
}

func (s *section) Exist(uid int) bool {
	if _, ok := s.users[uid]; ok {
		return true
	}
	return false
}

func (s *section) GetSeq(id int) (_seq seq.Sequence, err error) {
	if u, ok := s.users[id]; ok {
		u.curSeq++
		if u.curSeq > s.maxSeq {
			s.maxSeq += seq.Sequence(s.step)
		}
		return u.curSeq, nil
	}
	return 0, fmt.Errorf("cannot find user[%s]", id)
}

func (s *section) GetMaxSeq() seq.Sequence {
	return s.maxSeq
}

type user struct {
	uid    int
	curSeq seq.Sequence
}

func (u *user) GetUserSeq() seq.UserSequence {
	var re seq.UserSequence
	panic("todo")
	//todo: format user and curseq to UserSequnce
	return re
}