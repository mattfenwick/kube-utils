package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"sort"
)

type Diffs struct {
	Elements []*JDiff
}

func (d *Diffs) Add(e *JDiff) {
	d.Elements = append(d.Elements, e)
}

type JDiff struct {
	Type DiffType
	Path []string
	Old  interface{}
	New  interface{}
}

func JsonDiff(a interface{}, b interface{}, path []string, diffs *Diffs) {
	if a == nil && b != nil {
		diffs.Add(&JDiff{Type: DiffTypeAdd, Old: a, New: b, Path: path})
	} else if b == nil {
		diffs.Add(&JDiff{Type: DiffTypeRemove, Old: a, New: b, Path: path})
	} else {
		switch aVal := a.(type) {
		case map[string]interface{}:
			switch bVal := b.(type) {
			case map[string]interface{}:
				aKeys := MapKeys(aVal)
				sort.Strings(aKeys)
				for _, k := range aKeys {
					JsonDiff(aVal[k], bVal[k], append(path, fmt.Sprintf(`%s`, k)), diffs)
				}
				bKeys := MapKeys(bVal)
				sort.Strings(bKeys)
				for _, k := range bKeys {
					if _, ok := aVal[k]; !ok {
						diffs.Add(&JDiff{Type: DiffTypeAdd, New: bVal[k], Path: append(path, fmt.Sprintf(`%s`, k))})
					}
				}
			default:
				diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case []interface{}:
			switch bVal := b.(type) {
			case []interface{}:
				minLength := len(aVal)
				if len(bVal) < minLength {
					minLength = len(bVal)
				}
				for i, aSub := range aVal {
					newPath := append(path, fmt.Sprintf("%d", i))
					if i >= len(aVal) {
						diffs.Add(&JDiff{Type: DiffTypeAdd, New: bVal[i], Path: newPath})
					} else if i >= len(bVal) {
						diffs.Add(&JDiff{Type: DiffTypeRemove, Old: aSub, Path: newPath})
					} else {
						JsonDiff(aSub, bVal[i], newPath, diffs)
					}
				}
			default:
				diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case int:
			switch bVal := b.(type) {
			case int:
				if aVal != bVal {
					diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
				}
			default:
				diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case string:
			switch bVal := b.(type) {
			case string:
				if aVal != bVal {
					diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
				}
			default:
				diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case bool:
			switch bVal := b.(type) {
			case bool:
				if aVal != bVal {
					diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
				}
			default:
				diffs.Add(&JDiff{Type: DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		//case types.Nil: // TODO is this necessary?
		default:
			panic(errors.Errorf("unrecognized type: %T", aVal))
		}
	}
}