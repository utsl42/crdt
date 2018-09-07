package crdt

import "encoding/json"

// Gset is a grow-only set.
type GSet map[interface{}]struct{}

var (
	// GSet should implement the set interface.
	_ Set = GSet{}
)

// NewGSet returns an instance of GSet.
func NewGSet() GSet {
	return GSet{}
}

// Add lets you add an element to grow-only set.
func (g GSet) Add(elem interface{}) {
	g[elem] = struct{}{}
}

// Contains returns true if an element exists within the
// set or false otherwise.
func (g GSet) Contains(elem interface{}) bool {
	_, ok := g[elem]
	return ok
}

// Len returns the no. of elements present within GSet.
func (g GSet) Len() int {
	return len(g)
}

// Elems returns all the elements present in the set.
func (g GSet) Elems() []interface{} {
	elems := make([]interface{}, 0, len(g))

	for elem := range g {
		elems = append(elems, elem)
	}

	return elems
}

// MarshalJSON will be used to generate a serialized output
// of a given GSet.
func (g GSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Elems())
}

// UnmarshalJSON will be used to generate a serialized output
// of a given GSet.
func (g *GSet) UnmarshalJSON(b []byte) error {
	// if the map is nil, we have to create one
	if *g == nil {
		*g = make(map[interface{}]struct{})
	}

	var gsj []interface{}
	if err := json.Unmarshal(b, &gsj); err != nil {
		return err
	}
	for _, e := range gsj {
		g.Add(e)
	}
	return nil
}
