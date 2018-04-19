package base

import (
	"fmt"
	"github.com/dming/lodos/utils/uuid"
	"github.com/dming/lodos/addon/seq"
)

func NewSection(_flag int, _maxSeq seq.Sequence, _step seq.Sequence) (*section, error) {
	if _step <= 0 || _maxSeq <= 0 {
		return nil, fmt.Errorf("step or max_seq cannot equal or negative than 0")
	}
	return &section{
		flag: _flag,
		users: make(map[uuid.UUID]*user),
		maxSeq: _maxSeq,
		step: _step,
	}, nil
}

type section struct {
	flag int
	users map[uuid.UUID]*user
	maxSeq seq.Sequence
	step seq.Sequence
}

func (s *section) GetFlag() int {
	return s.flag
}

func (s *section) AddUser(id uuid.UUID, _seq seq.Sequence) error {
	if _, ok := s.users[id]; ok {
		return fmt.Errorf("user[%s] already exist")
	}
	s.users[id] = &user {
		uid:    id,
		curSeq: _seq,
	}
	return nil
}

func (s *section) Exist(uid uuid.UUID) bool {
	if _, ok := s.users[uid]; ok {
		return true
	}
	return false
}

func (s *section) GetSeq(id uuid.UUID) (_seq seq.Sequence, err error) {
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
	uid    uuid.UUID
	curSeq seq.Sequence
}

func (u *user) GetUserSeq() seq.UserSequence {
	var re seq.UserSequence
	panic("todo")
	//todo: format user and curseq to UserSequnce
	return re
}