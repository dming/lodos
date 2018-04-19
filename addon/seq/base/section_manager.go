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
	sections map[int]*section
}

func (sm *section_manager) AddSection (sec *section) error {
	if _, ok := sm.sections[sec.flag]; ok {
		return fmt.Errorf("section[%d] already exist", sec.flag)
	}
	sm.sections[sec.flag] = sec
	return nil
}

func (sm *section_manager) GetSection (flag int) *section {
	if s, ok := sm.sections[flag]; ok {
		return s
	}
	return nil
}

