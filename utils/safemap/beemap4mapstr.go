// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safemap

import (
	"sync"
)

// BeeMap is a map with lock
type BeeMap4MapStr struct {
	lock *sync.RWMutex
	bm   map[string]*BeeMap4String
}

// NewBeeMap return new safemap
func NewBeeMap4MapStr() *BeeMap4MapStr {
	return &BeeMap4MapStr{
		lock: new(sync.RWMutex),
		bm:   map[string]*BeeMap4String{},
	}
}

// Get from maps return the k's value
func (m *BeeMap4MapStr) Get(k string) *BeeMap4String {
	m.lock.RLock()
	//defer m.lock.RUnlock()

	if val, ok := m.bm[k]; ok {
		m.lock.RUnlock()
		return val
	}
	m.lock.RUnlock()
	return nil
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *BeeMap4MapStr) Set(k string, v *BeeMap4String) {
	m.lock.Lock()
	m.bm[k] = v
	m.lock.Unlock()
}

// Check Returns true if k is exist in the map.
func (m *BeeMap4MapStr) Check(k string) bool {
	m.lock.RLock()
	//defer m.lock.RUnlock()

	if _, ok := m.bm[k]; ok {
		m.lock.RUnlock()
		return true
	}
	m.lock.RUnlock()
	return false
}

// Delete the given key and value.
func (m *BeeMap4MapStr) Delete(k string) {
	m.lock.Lock()
	//defer m.lock.Unlock()

	delete(m.bm, k)
	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *BeeMap4MapStr) Items() map[string]map[string]string {
	m.lock.RLock()
	//defer m.lock.RUnlock()

	r := make(map[string]map[string]string)
	for k, v := range m.bm {
		r[k] = v.bm
	}
	m.lock.RUnlock()
	return r
}
