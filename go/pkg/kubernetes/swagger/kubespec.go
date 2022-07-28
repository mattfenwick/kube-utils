package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/base"
	"github.com/mattfenwick/collections/pkg/function"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"strings"
)

type SpecType struct {
	AdditionalProperties        *SpecType                `json:"additionalProperties,omitempty"`
	Description                 string                   `json:"description,omitempty"`
	Format                      string                   `json:"format,omitempty"`
	Items                       *SpecType                `json:"items,omitempty"`
	Properties                  map[string]*SpecType     `json:"properties,omitempty"`
	Ref                         string                   `json:"$ref,omitempty"`
	Required                    []string                 `json:"required,omitempty"`
	Type                        string                   `json:"type,omitempty"`
	XKubernetesListMapKeys      []string                 `json:"x-kubernetes-list-map-keys,omitempty"`
	XKubernetesListType         string                   `json:"x-kubernetes-list-type,omitempty"`
	XKubernetesPatchMergeKey    string                   `json:"x-kubernetes-patch-merge-key,omitempty"`
	XKubernetesPatchStrategy    string                   `json:"x-kubernetes-patch-strategy,omitempty"`
	XKubernetesGroupVersionKind []*GVK                   `json:"x-kubernetes-group-version-kind,omitempty"`
	XKubernetesUnions           []map[string]interface{} `json:"x-kubernetes-unions,omitempty"`
}

type KubeSpec struct {
	Definitions map[string]*SpecType `json:"definitions"`
	Info        struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}

func enforceInvariant(specType *SpecType) {
	counts := slice.Filter(function.Id[bool], []bool{specType.Ref != "", specType.Type != ""})
	if len(counts) != 1 && specType.Description == "" {
		logrus.Errorf("INVARIANT violated: %d; %+v", len(counts), specType)
	}
}

func (s *KubeSpec) MustGetDefinition(name string) *SpecType {
	val, ok := s.Definitions[name]
	if !ok {
		panic(errors.Errorf("unable to find definition for %s", name))
	}
	return val
}

func (s *KubeSpec) VisitSpecType(resolvedTypes map[string]*ResolvedType, path Path, specType *SpecType, visit func(path Path, resolved *ResolvedType, circular string)) *ResolvedType {
	enforceInvariant(specType)

	// visit AFTER processing of the type is done
	var resolved *ResolvedType
	defer visit(path, resolved, "")

	if specType.Ref != "" {
		refName := ParseRef(specType.Ref)
		newPath := path.Append(SpecPath{Ref: true})
		if resolvedTypes[refName] != nil {
			// done: NOT circular
			visit(newPath, resolvedTypes[refName], "")
			resolved = resolvedTypes[refName]
		} else if _, ok := resolvedTypes[refName]; ok {
			// in progress: circular
			visit(newPath, nil, refName)
			resolved = &ResolvedType{Circular: refName}
		} else {
			// hasn't been seen yet
			resolvedTypes[refName] = nil
			resolved = s.VisitSpecType(resolvedTypes, newPath, s.MustGetDefinition(refName), visit)
			resolvedTypes[refName] = resolved
		}
	} else {
		switch specType.Type {
		case "":
			logrus.Debugf("skipping empty type: %+v", strings.Join(path.ToStringPieces(), "."))
			resolved = &ResolvedType{Empty: true}
		case "array":
			resolved = &ResolvedType{Array: s.VisitSpecType(resolvedTypes, path.Append(SpecPath{Array: true}), specType.Items, visit)}
		case "object":
			obj := &ResolvedObject{Properties: map[string]*ResolvedType{}}
			for propName, prop := range specType.Properties {
				obj.Properties[propName] = s.VisitSpecType(resolvedTypes, path.Append(SpecPath{ObjectProperty: true}).Append(SpecPath{FieldAccess: propName}), prop, visit)
			}
			if specType.AdditionalProperties != nil {
				obj.AdditionalProperties = s.VisitSpecType(resolvedTypes, path.Append(SpecPath{FieldAccess: "additionalProperties"}), specType.AdditionalProperties, visit)
			}
			resolved = &ResolvedType{Object: obj}
		case "boolean", "string", "integer", "number":
			resolved = &ResolvedType{Primitive: specType.Type}
			logrus.Debugf("found primitive: %s", specType.Type)
		default:
			panic(errors.Errorf("TODO unsupported type %s: %+v, %+v", specType.Type, path, specType))
		}
	}
	return resolved
}

func (s *KubeSpec) Visit(visit func(path Path, resolved *ResolvedType, circular string)) (map[string]*ResolvedType, map[string]map[string]*ResolvedType) {
	resolvedTypes := map[string]*ResolvedType{}
	for defName, def := range s.Definitions {
		resolvedTypes[defName] = nil
		resolved := s.VisitSpecType(resolvedTypes, []SpecPath{{FieldAccess: defName}}, def, visit)
		resolvedTypes[defName] = resolved
	}
	gvks := map[string]map[string]*ResolvedType{}
	for gvkString, resolved := range resolvedTypes {
		gvk := ParseGVK(gvkString)
		if _, ok := gvks[gvk.Kind]; !ok {
			gvks[gvk.Kind] = map[string]*ResolvedType{}
		}
		gvks[gvk.Kind][gvk.GroupVersion()] = resolved
	}
	return resolvedTypes, gvks
}

type SpecPath struct {
	FieldAccess    string
	Ref            bool
	ObjectProperty bool
	Array          bool
}

type Path []SpecPath

func (p Path) Append(piece SpecPath) Path {
	return slice.Append(p, []SpecPath{piece})
}

func (p Path) ToStringPieces() []string {
	var elems []string
	for _, piece := range p {
		if piece.FieldAccess != "" {
			elems = append(elems, piece.FieldAccess)
		} else if piece.Ref {
			// nothing to do
		} else if piece.Array {
			elems = append(elems, "[]")
		} else if piece.ObjectProperty {
			// nothing to do ??  TODO decide
		} else {
			panic(errors.Errorf("invalid SpecPath value: %+v", piece))
		}
	}
	return elems
}

func (s *KubeSpec) ResolveStructure(args *ExplainResourceArgs) {
	_, gvks := s.Visit(func(path Path, resolved *ResolvedType, circular string) {
		if circular == "" {
			fmt.Printf("%+v -- %+v\n", path.ToStringPieces(), resolved)
		} else {
			fmt.Printf("%+v\n  CIRCULAR %s\n", path.ToStringPieces(), circular)
		}
	})
	resources := set.NewSet(args.TypeNames)
	fmt.Printf("\n\n\n\n")
	for _, name := range slice.Sort(maps.Keys(gvks)) {
		if len(args.TypeNames) > 0 && !resources.Contains(name) {
			continue
		}
		switch args.Format {
		case "debug":
			fmt.Printf("%s:\n", name)
			for gv, kind := range gvks[name] {
				fmt.Printf("gv: %s:\n", gv)
				for _, path := range kind.Paths([]string{name}) {
					if args.Depth == 0 || len(path.Fst) <= args.Depth {
						fmt.Printf("  %+v: %s\n", path.Fst, path.Snd)
					}
				}
			}
			//json.Print(resolved[name])
			fmt.Printf("\n\n")
		case "table":
			panic("TODO")
		case "condensed":
			fmt.Printf("%s:\n", name)
			for gv, kind := range gvks[name] {
				fmt.Printf("%s:\n", gv)
				for _, path := range kind.Paths([]string{name}) {
					if args.Depth == 0 || len(path.Fst) <= args.Depth {
						prefix := strings.Repeat("  ", len(path.Fst)-1)
						typeString := fmt.Sprintf("%s%s", prefix, path.Fst[len(path.Fst)-1])
						//fmt.Printf("%s\n", strings.Join(path.Fst, "."))
						fmt.Printf("%-60s    %s\n", typeString, path.Snd)
					}
				}
				//json.Print(resolved[name])
				fmt.Printf("\n\n")
			}
		default:
			panic(errors.Errorf("invalid output format: %s", args.Format))
		}
	}
}

//func (s *KubeSpec) ResolveGVKs() {
//	gvksByResource := map[string]map[string]*SpecType{}
//	s.Visit(func(path Path, resolved *ResolvedType, circular string) {
//		if circular == "" {
//			fmt.Printf("%+v -- %s\n", path.ToStringPieces(), resolved)
//		} else {
//			fmt.Printf("%+v\n  CIRCULAR %s\n", path.ToStringPieces(), circular)
//		}
//	})
//}

// TODO distinguish between gvk and parsed name
//type ResolvedGVK struct {
//	GVK *GVK
//	Name string
//	Type *ResolvedType
//}

type ResolvedObject struct {
	Properties           map[string]*ResolvedType
	AdditionalProperties *ResolvedType
}

type ResolvedType struct {
	Empty     bool
	Primitive string
	Array     *ResolvedType
	Object    *ResolvedObject
	Circular  string
}

func (r *ResolvedType) Paths(context []string) []*base.Pair[[]string, string] {
	if r.Empty {
		return []*base.Pair[[]string, string]{base.NewPair(utils.CopySlice(context), "?")}
	} else if r.Primitive != "" {
		return []*base.Pair[[]string, string]{base.NewPair(utils.CopySlice(context), r.Primitive)}
	} else if r.Array != nil {
		return append(
			[]*base.Pair[[]string, string]{base.NewPair(utils.CopySlice(context), "[]")},
			r.Array.Paths(slice.Append(context, []string{"[]"}))...)
	} else if r.Object != nil {
		out := []*base.Pair[[]string, string]{
			base.NewPair(utils.CopySlice(context), "object"),
		}
		for _, name := range slice.Sort(maps.Keys(r.Object.Properties)) {
			prop := r.Object.Properties[name]
			out = append(out, prop.Paths(slice.Append(context, []string{name}))...)
		}
		if r.Object.AdditionalProperties != nil {
			out = append(out, r.Object.AdditionalProperties.Paths(slice.Append(context, []string{"additionalProperties"}))...)
		}
		return out
	} else if r.Circular != "" {
		return []*base.Pair[[]string, string]{base.NewPair(utils.CopySlice(context), "circular: "+r.Circular)}
	}
	panic(errors.Errorf("invalid ResolvedType value: %+v", r))
}
