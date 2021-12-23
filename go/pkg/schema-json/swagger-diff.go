package schema_json

import (
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
)

// We care about:
// - type definition added or deleted
// - otherwise: dig into type definition, see if property added or deleted

//type SwaggerAdditionalPropertyDiff struct {
//	Attributes *utils.MapDiff
//	Items      *utils.MapDiff
//}

//func (s *SwaggerAdditionalProperty) Dict() map[string]string {
//	return utils.MapFilterEmptyValues(map[string]string{
//		"format": s.Format,
//		"ref":    s.Ref,
//		"type":   s.Type,
//	})
//}

//func (s *SwaggerAdditionalProperty) Diff(other *SwaggerAdditionalProperty) *SwaggerAdditionalPropertyDiff {
//	// TODO handle nils in either side
//	return &SwaggerAdditionalPropertyDiff{
//		Items:      utils.CompareMaps(s.Items, other.Items),
//		Attributes: utils.CompareMaps(s.Dict(), other.Dict()),
//	}
//}

type Change struct {
	FieldName string
	Old       interface{}
	New       interface{}
	MapDiff   *utils.MapDiff
}

func (s *SwaggerAdditionalProperty) Compare(other *SwaggerAdditionalProperty) []*Change {
	var changes []*Change
	if s.Format != other.Format {
		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
	}
	if s.Ref != other.Ref {
		changes = append(changes, &Change{FieldName: "Ref", Old: s.Ref, New: other.Ref})
	}
	if s.Type != other.Type {
		changes = append(changes, &Change{FieldName: "Type", Old: s.Type, New: other.Type})
	}
	itemsDiff := utils.CompareMaps(s.Items, other.Items)
	if !itemsDiff.IsSame() {
		changes = append(changes, &Change{FieldName: "Items", Old: s.Items, New: other.Items})
	}
	return changes
}

//func (s *SwaggerProperty) Compare(other *SwaggerProperty) []*Change {
//	var changes []*Change
//
//	//if s.AdditionalProperties == nil && other.AdditionalProperties == nil {
//	//
//	//} else if s.AdditionalProperties == nil {
//	//
//	//} else if other.AdditionalProperties == nil {
//	//
//	//} else {
//	//
//	//}
//
//	if s.Description != other.Description {
//		changes = append(changes, &Change{FieldName: "Description", Old: s.Description, New: other.Description})
//	}
//	if s.Format != other.Format {
//		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
//	}
//	if s.Ref != other.Ref {
//		changes = append(changes, &Change{FieldName: "Ref", Old: s.Ref, New: other.Ref})
//	}
//	if s.Type != other.Type {
//		changes = append(changes, &Change{FieldName: "Type", Old: s.Type, New: other.Type})
//	}
//	itemsDiff := utils.CompareMaps(s.Items, other.Items)
//	if !itemsDiff.IsSame() {
//		changes = append(changes, &Change{FieldName: "Items", Old: s.Items, New: other.Items, MapDiff: itemsDiff})
//	}
//	return changes
//}

//type SwaggerPropertyDiff struct {
//	AdditionalProperties *SwaggerAdditionalPropertyDiff
//	Attributes           *utils.MapDiff
//	Items                *utils.MapDiff
//}

//func (s *SwaggerProperty) Dict() map[string]string {
//	return utils.MapFilterEmptyValues(map[string]string{
//		"description": s.Description,
//		"format":      s.Format,
//		"ref":         s.Ref,
//		"type":        s.Type,
//	})
//}

//func (s *SwaggerProperty) Diff(other *SwaggerProperty) *SwaggerPropertyDiff {
//	return &SwaggerPropertyDiff{
//		AdditionalProperties: s.AdditionalProperties.Diff(other.AdditionalProperties),
//		Items:                utils.CompareMaps(s.Items, other.Items),
//		Attributes:           utils.CompareMaps(s.Dict(), other.Dict()),
//	}
//}

//func (s *SwaggerDefinition) Dict() map[string]string {
//	return utils.MapFilterEmptyValues(map[string]string{
//		"description": s.Description,
//		"format":      s.Format,
//		"type":        s.Type,
//	})
//}
//
//type SwaggerDefinitionDiff struct {
//	Attributes *utils.MapDiff
//	Properties int // TODO what's the type here? *utils.MapDiff
//}

//func (s *SwaggerDefinition) Diff(other *SwaggerDefinition) []*Change {
//	var changes []*Change
//	if s.Description != other.Description {
//		changes = append(changes, &Change{FieldName: "Description", Old: s.Description, New: other.Description})
//	}
//	if s.Format != other.Format {
//		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
//	}
//	if s.Type != other.Type {
//		changes = append(changes, &Change{FieldName: "Type", Old: s.Type, New: other.Type})
//	}
//
//	for k, aVal := range s.Properties {
//		bVal, ok := other.Properties[k]
//		if ok && SwaggerPropertyIsSame(aVal, bVal) {
//			// nothing to do
//		} else if ok {
//			// change
//		} else {
//			changes = append(changes, &Change{FieldName: fmt.Sprintf("Properties.%s", k), Old: aVal, New: bVal})
//		}
//	}
//	for k := range other.Properties {
//		if _, ok := s.Properties[k]; !ok {
//			// added
//		}
//	}
//	itemsDiff := utils.CompareMaps(s.Items, other.Items)
//	if !itemsDiff.IsSame() {
//		changes = append(changes, &Change{FieldName: "Items", Old: s.Items, New: other.Items, MapDiff: itemsDiff})
//	}
//	return changes
//}

//func (s *SwaggerSpec) Diff(other *SwaggerSpec) error {
//	panic("TODO")
//}
