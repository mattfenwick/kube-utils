package schema_json

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/pkg/errors"
)

type PathComponent struct {
	ArrayIndex *int
	MapKey     *string
	MapValue   *string
}

func NewArrayPathComponent(index int) *PathComponent {
	return &PathComponent{ArrayIndex: &index}
}

func NewMapKeyPathComponent(key string) *PathComponent {
	return &PathComponent{MapKey: &key}
}

func NewMapValuePathComponent(key string) *PathComponent {
	return &PathComponent{MapValue: &key}
}

func (p *PathComponent) RawString() string {
	if p.MapKey != nil {
		return *p.MapKey
	} else if p.MapValue != nil {
		return *p.MapValue
	} else if p.ArrayIndex != nil {
		return fmt.Sprintf("%d", *p.ArrayIndex)
	} else {
		simulator.DoOrDie(errors.Errorf("invalid PathComponent: %+v", p))
	}
	panic(errors.Errorf("this shouldn't happen"))
}

func (p *PathComponent) PathString() string {
	if p.MapKey != nil {
		return fmt.Sprintf(`{"%s"}`, *p.MapKey)
	} else if p.MapValue != nil {
		return fmt.Sprintf(`["%s"]`, *p.MapValue)
	} else if p.ArrayIndex != nil {
		return fmt.Sprintf(`[%d]`, *p.ArrayIndex)
	} else {
		simulator.DoOrDie(errors.Errorf("invalid PathComponent: %+v", p))
	}
	panic(errors.Errorf("this shouldn't happen"))
}

func PathString(components []*PathComponent) []string {
	path := make([]string, len(components))
	for i, component := range components {
		path[i] = component.PathString()
	}
	return path
}
