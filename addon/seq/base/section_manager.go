package base

import "fmt"

var sectionManager *section_manager

func GetSectionManager() *section_manager {
	if sectionManager == nil {
		return &section_manager{
			sections: make(map[int]*section),
		}
	}
	return sectionManager
}

type section_manager struct {
	startId int
	size int
	sections []*section //len(sections) == size
}

func (sm *section_manager) AddSection (sec *section) error {
	if _, ok := sm.sections[sec.begin_id]; ok {
		return fmt.Errorf("section[%d] already exist", sec.begin_id)
	}
	sm.sections[sec.begin_id] = sec
	return nil
}

func (sm *section_manager) GetSection (flag int) *section {
	if s, ok := sm.sections[flag]; ok {
		return s
	}
	return nil
}

