package schema_json

type SwaggerProperty struct {
	Description string
	Items       map[string]*SwaggerProperty
	Type        string
	Ref         string `json:"$ref"`
}

type SwaggerSpec struct {
	Definitions map[string]struct {
		Description                 string
		Properties                  map[string]*SwaggerProperty
		Type                        string
		XKubernetesGroupVersionKind []struct {
			Group   string
			Kind    string
			Version string
		} `json:"x-kubernetes-group-version-kind"`
	}
	Info struct {
		Title   string
		Version string
	}
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}
