package json_traversal

func JsonFindByPath(obj interface{}, path []*PathComponent) interface{} {
	for _, component := range path {
		switch o := obj.(type) {
		case []interface{}:
			if component.ArrayIndex != nil {
				if len(o) > *component.ArrayIndex {
					obj = o[*component.ArrayIndex]
				} else {
					return nil
				}
			} else {
				return nil
			}
		case map[string]interface{}:
			if component.MapKey != nil {
				if _, ok := o[*component.MapKey]; ok {
					obj = *component.MapKey
				} else {
					return nil
				}
			} else if component.MapValue != nil {
				obj = o[*component.MapValue]
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return obj
}
