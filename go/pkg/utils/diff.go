package utils

type DiffType = string

const (
	DiffTypeAdd    DiffType = "DiffTypeAdd"
	DiffTypeRemove DiffType = "DiffTypeRemove"
	DiffTypeChange DiffType = "DiffTypeChange"
	DiffTypeSame   DiffType = "DiffTypeSame"
)

type Diff struct {
	Type     DiffType
	Previous *string
	Current  *string
}

type MapDiff struct {
	Elements map[string]*Diff
}

func (m *MapDiff) IsSame() bool {
	for _, v := range m.Elements {
		if v.Type != DiffTypeSame {
			return false
		}
	}
	return true
}

func CompareMaps(as map[string]string, bs map[string]string) *MapDiff {
	mapDiff := &MapDiff{Elements: map[string]*Diff{}}
	for k, aVal := range as {
		bVal, ok := bs[k]
		mapDiff.Elements[k] = &Diff{Previous: Pointer(aVal), Current: &bVal}
		if ok && aVal == bVal {
			mapDiff.Elements[k].Type = DiffTypeSame
		} else if ok {
			mapDiff.Elements[k].Type = DiffTypeChange
		} else {
			mapDiff.Elements[k].Type = DiffTypeRemove
		}
	}
	for k, bVal := range bs {
		if _, ok := as[k]; !ok {
			mapDiff.Elements[k] = &Diff{Type: DiffTypeAdd, Previous: nil, Current: Pointer(bVal)}
		}
	}
	return mapDiff
}

func Pointer(s string) *string {
	return &s
}

/*
type MapDiff struct {
	Added   map[string]interface{}
	Removed map[string]interface{}
	Same map[string]interface{}
	Changed    map[string]*Diff
}

func CompareMaps(a map[string]interface{}, b map[string]interface{}) *MapDiff {
	diff := &MapDiff{
		Added: map[string]interface{}{},
		Removed: map[string]interface{}{},
		Same: map[string]interface{}{},
		Changed: map[string]*Diff{},
	}
	for k, v := range a {

	}
	return diff
}

// Types:
// - dict
// - slice
// - number (int?)
// - string
// - bool
// - nil

type Diff struct {
	Type DiffType
	Old interface{}
	New interface{}
}

func CompareValues(a interface{}, b interface{}) *Diff {
	if a == nil {
		return &Diff{Type: DiffTypeAdd, New:  b}
	} else if b == nil {
		return &Diff{Type: DiffTypeRemove, Old:  a}
	} else {
		switch aVal := a.(type) {
		case map[string]interface{}:
			switch bVal := b.(type) {
			case map[string]interface{}:
				// TODO
			default:
				return &Diff{Type: DiffTypeChange, Old: aVal, New: bVal}
			}
		case []interface{}:
			// TODO
		case int:
			switch bVal := b.(type) {
			case int:
				diffType := DiffTypeChange
				if aVal == bVal {
					diffType = DiffTypeSame
				}
				return &Diff{Type: diffType, Old: aVal, New: bVal}
			}
		case string:
			switch bVal := b.(type) {
			case string:
				diffType := DiffTypeChange
				if aVal == bVal {
					diffType = DiffTypeSame
				}
				return &Diff{Type: diffType, Old: aVal, New: bVal}
			}
		case bool:
			switch bVal := b.(type) {
			case bool:
				diffType := DiffTypeChange
				if aVal == bVal {
					diffType = DiffTypeSame
				}
				return &Diff{Type: diffType, Old: aVal, New: bVal}
			}
		//case types.Nil: // TODO is this necessary?
		default:
			panic(errors.Errorf("unrecognized type: %T", aVal))
		}
	}
}
*/
