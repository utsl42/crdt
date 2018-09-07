package crdt

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

type ORSet struct {
	addMap map[string]GSet
	rmMap  map[string]GSet
}

func NewORSet() *ORSet {
	return &ORSet{
		addMap: make(map[string]GSet),
		rmMap:  make(map[string]GSet),
	}
}

func (o *ORSet) Add(value string) {
	newID := uuid.NewV4().String()
	if m, ok := o.addMap[value]; ok {
		m.Add(newID)
		o.addMap[value] = m
		return
	}

	m := NewGSet()
	m.Add(newID)
	o.addMap[value] = m
}

func (o *ORSet) Remove(value string) {
	r, ok := o.rmMap[value]
	if !ok {
		r = NewGSet()
	}

	if m, ok := o.addMap[value]; ok {
		for _, uid := range m.Elems() {
			r.Add(uid)
		}
	}

	o.rmMap[value] = r
}

func (o *ORSet) Contains(value string) bool {
	addMap, ok := o.addMap[value]
	if !ok {
		return false
	}

	rmMap, ok := o.rmMap[value]
	if !ok {
		return true
	}

	for _, uid := range addMap.Elems() {
		if ok := rmMap.Contains(uid); !ok {
			return true
		}
	}

	return false
}

func (o *ORSet) Merge(r *ORSet) {
	for value, m := range r.addMap {
		addMap, ok := o.addMap[value]
		if ok {
			for _, uid := range m.Elems() {
				addMap.Add(uid)
			}

			continue
		}

		o.addMap[value] = m
	}

	for value, m := range r.rmMap {
		rmMap, ok := o.rmMap[value]
		if ok {
			for _, uid := range m.Elems() {
				rmMap.Add(uid)
			}

			continue
		}

		o.rmMap[value] = m
	}
}

func (o *ORSet) Elems() []string {
	var e []string
	for k := range o.addMap {
		if o.Contains(k) {
			e = append(e, k)
		}
	}
	return e
}

type orsetJSON struct {
	AddMap map[string]GSet
	RmMap  map[string]GSet
}

func (o *ORSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(&orsetJSON{
		AddMap: o.addMap,
		RmMap:  o.rmMap,
	})
}

func (o *ORSet) UnmarshalJSON(b []byte) error {
	v := &orsetJSON{
		AddMap: make(map[string]GSet),
		RmMap:  make(map[string]GSet),
	}
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	o.addMap = v.AddMap
	o.rmMap = v.RmMap

	return nil
}
