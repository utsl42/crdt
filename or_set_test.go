package crdt

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/satori/go.uuid"
)

func TestORSetAddContains(t *testing.T) {
	orSet := NewORSet()

	var testValue string = "object"

	if orSet.Contains(testValue) {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}

	orSet.Add(testValue)

	if !orSet.Contains(testValue) {
		t.Errorf("Expected set to contain: %v, but not found", testValue)
	}
}

func TestORSetAddRemoveContains(t *testing.T) {
	orSet := NewORSet()

	var testValue string = "object"
	orSet.Add(testValue)

	orSet.Remove(testValue)

	if orSet.Contains(testValue) {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

func TestORSetAddRemoveAddContains(t *testing.T) {
	orSet := NewORSet()

	var testValue string = "object"

	orSet.Add(testValue)
	orSet.Remove(testValue)
	orSet.Add(testValue)

	if !orSet.Contains(testValue) {
		t.Errorf("Expected set to contain: %v, but not found", testValue)
	}
}

func TestORSetAddAddRemoveContains(t *testing.T) {
	orSet := NewORSet()

	var testValue string = "object"

	orSet.Add(testValue)
	orSet.Add(testValue)
	orSet.Remove(testValue)

	if orSet.Contains(testValue) {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

func TestORSetMerge(t *testing.T) {
	type addRm struct {
		addSet []string
		rmSet  []string
	}

	for _, tt := range []struct {
		setOne  addRm
		setTwo  addRm
		valid   map[string]struct{}
		invalid map[string]struct{}
	}{
		{
			addRm{[]string{"object1"}, []string{}},
			addRm{[]string{}, []string{"object1"}},
			map[string]struct{}{
				"object1": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{}, []string{"object1"}},
			addRm{[]string{"object1"}, []string{}},
			map[string]struct{}{
				"object1": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{"object1"}, []string{"object1"}},
			addRm{[]string{}, []string{}},
			map[string]struct{}{},
			map[string]struct{}{
				"object1": struct{}{},
			},
		},
		{
			addRm{[]string{}, []string{}},
			addRm{[]string{"object1"}, []string{"object1"}},
			map[string]struct{}{},
			map[string]struct{}{
				"object1": struct{}{},
			},
		},
		{
			addRm{[]string{"object2"}, []string{"object1"}},
			addRm{[]string{"object1"}, []string{"object2"}},
			map[string]struct{}{
				"object1": struct{}{},
				"object2": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{"object2", "object1"}, []string{"object1"}},
			addRm{[]string{"object1", "object2"}, []string{"object2"}},
			map[string]struct{}{
				"object1": struct{}{},
				"object2": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{"object2", "object1"}, []string{"object1", "object2"}},
			addRm{[]string{"object1", "object2"}, []string{"object2", "object1"}},
			map[string]struct{}{},
			map[string]struct{}{
				"object1": struct{}{},
				"object2": struct{}{},
			},
		},
	} {
		orset1, orset2 := NewORSet(), NewORSet()

		for _, add := range tt.setOne.addSet {
			orset1.Add(add)
		}

		for _, rm := range tt.setOne.rmSet {
			orset1.Remove(rm)
		}

		for _, add := range tt.setTwo.addSet {
			orset2.Add(add)
		}

		for _, rm := range tt.setTwo.rmSet {
			orset2.Remove(rm)
		}

		orset1.Merge(orset2)

		for obj, _ := range tt.valid {
			if !orset1.Contains(obj) {
				t.Errorf("expected set to contain: %v", obj)
			}
		}

		for obj, _ := range tt.invalid {
			if orset1.Contains(obj) {
				t.Errorf("expected set to not contain: %v", obj)
			}
		}
	}
}

func TestORSetElems(t *testing.T) {
	for _, tt := range []struct {
		add []string
		rem []string
	}{
		{[]string{}, []string{}},
		{[]string{"1"}, []string{}},
		{[]string{"1", "2", "3"}, []string{"2"}},
		{[]string{"1", "100", "1000", "-1"}, []string{"100"}},
		{[]string{"alpha", "beta", "1", "2"}, []string{"beta", "2"}},
	} {
		orset := NewORSet()

		expectedElems := map[string]struct{}{}
		for _, i := range tt.add {
			expectedElems[i] = struct{}{}
			orset.Add(i)
		}

		for _, i := range tt.rem {
			delete(expectedElems, i)
			orset.Remove(i)
		}

		actualElems := map[string]struct{}{}
		for _, i := range orset.Elems() {
			actualElems[i] = struct{}{}
		}

		if !reflect.DeepEqual(expectedElems, actualElems) {
			t.Errorf("expected set to contain: %v, actual: %v", expectedElems, actualElems)
		}
	}
}

func TestORSetTorture(t *testing.T) {
	orset1 := NewORSet()
	added := NewGSet()
	for i := 1; i < 1000; i++ {
		uid := uuid.NewV4().String()
		orset1.Add(uid)
		added.Add(uid)
	}

	buf, err := json.Marshal(orset1)
	if err != nil {
		t.Error(err)
	}

	removed1 := NewGSet()
	for i, k := range added.Elems() {
		if i%5 == 1 {
			orset1.Remove(k.(string))
			removed1.Add(k)
		}
	}

	orset2 := NewORSet()
	if err := json.Unmarshal(buf, &orset2); err != nil {
		t.Error(err)
	}

	removed2 := NewGSet()
	for i, k := range added.Elems() {
		if i%3 == 1 {
			orset2.Remove(k.(string))
			removed2.Add(k)
		}
	}

	orset1.Merge(orset2)
	orset2.Merge(orset1)

	for _, item := range orset1.Elems() {
		if removed1.Contains(item) || removed2.Contains(item) {
			t.Errorf("orset1 contains %v, expected it to be removed", item)
		}
		if !orset2.Contains(item) {
			t.Errorf("orset2 missing expected %v", item)
		}
	}
	for _, item := range orset2.Elems() {
		if removed1.Contains(item) || removed2.Contains(item) {
			t.Errorf("orset2 contains %v, expected it to be removed", item)
		}
		if !orset1.Contains(item) {
			t.Errorf("orset1 missing expected %v", item)
		}
	}
}
