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
type BeeMap4String struct {
	lock *sync.RWMutex
	bm   map[string]string
}

// NewBeeMap return new safemap
func NewBeeMap4String() *BeeMap4String {
	return &BeeMap4String{
		lock: new(sync.RWMutex),
		bm:   map[string]string{},
	}
}

// Get from maps return the k's value
func (m *BeeMap4String) Get(k string) string {
	m.lock.RLock()
	//defer m.lock.RUnlock()

	if val, ok := m.bm[k]; ok {
		m.lock.RUnlock()
		return val
	}
	m.lock.RUnlock()
	return ""
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *BeeMap4String) Set(k string, v string) {
	m.lock.Lock()
	m.bm[k] = v
	m.lock.Unlock()
}

// Check Returns true if k is exist in the map.
func (m *BeeMap4String) Check(k string) bool {
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
func (m *BeeMap4String) Delete(k string) {
	m.lock.Lock()
	//defer m.lock.Unlock()

	delete(m.bm, k)
	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *BeeMap4String) Items() map[string]string {
	m.lock.RLock()
	//defer m.lock.RUnlock()

	r := make(map[string]string)
	for k, v := range m.bm {
		r[k] = v
	}
	m.lock.RUnlock()
	return r
}
