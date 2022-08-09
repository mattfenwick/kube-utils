package vanilla

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/yaml"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var (
	ClusterScopedResources = set.FromSlice([]string{"ClusterRole", "ClusterRoleBinding", "CustomResourceDefinition"})
)

// Run
//
//	TODO why is this file here?  it doesn't seem to be used?
func Run(path string) {
	data, err := ioutil.ReadFile(path)
	utils.DoOrDie(errors.Wrapf(err, "unable to read file"))

	values, err := yaml.ParseMany[interface{}](data)
	for _, v := range values {
		fmt.Printf("found value: %+v\n\n", v)
	}
	utils.DoOrDie(err)

	resources, err := ParseResources(values)
	utils.DoOrDie(err)

	for kind, clusterResource := range resources.ClusterScoped {
		fmt.Printf("cluster-scoped %s:\n", kind)
		for name := range clusterResource {
			fmt.Printf("  %s\n", name)
		}
	}
	fmt.Println()

	for kind, kindResources := range resources.Namespaced {
		fmt.Printf("kind %s:\n", kind)
		for namespace, namespacedResources := range kindResources {
			fmt.Printf("  namespace %s\n", namespace)
			for name := range namespacedResources {
				fmt.Printf("    %s\n", name)
			}
		}
	}
}

type KubeResources struct {
	// Order: kind, name
	ClusterScoped map[string]map[string]*Node
	// Order: kind, namespace, name
	Namespaced map[string]map[string]map[string]*Node
}

func (k *KubeResources) Add(node *Node) error {
	if node.IsClusterScoped() {
		if _, ok := k.ClusterScoped[node.Kind]; !ok {
			k.ClusterScoped[node.Kind] = map[string]*Node{}
		}
		kinds := k.ClusterScoped[node.Kind]
		if _, ok := kinds[node.Name]; ok {
			return errors.Errorf("duplicate cluster-scoped resource: %s/%s", node.Kind, node.Name)
		}
		kinds[node.Name] = node
		return nil
	}
	if _, ok := k.Namespaced[node.Kind]; !ok {
		k.Namespaced[node.Kind] = map[string]map[string]*Node{}
	}
	kinds := k.Namespaced[node.Kind]
	if _, ok := kinds[node.ResolvedNamespace()]; !ok {
		kinds[node.ResolvedNamespace()] = map[string]*Node{}
	}
	ns := kinds[node.ResolvedNamespace()]
	if _, ok := ns[node.Name]; ok {
		return errors.Errorf("duplicate object: %s/%s/%s", node.Kind, node.ResolvedNamespace(), node.Name)
	}
	ns[node.Name] = node
	return nil
}

type Node struct {
	Kind       string
	Namespace  *string
	Name       string
	References []*Node
}

func NewNode(kind string, namespace string, name string) *Node {
	node := &Node{
		Kind:       kind,
		Namespace:  nil,
		Name:       name,
		References: nil,
	}
	if ClusterScopedResources.Contains(kind) {
		node.Namespace = &namespace
	}
	return node
}

func (n *Node) IsClusterScoped() bool {
	return n.Namespace == nil
}

func (n *Node) ResolvedNamespace() string {
	if n.IsClusterScoped() {
		err := errors.Errorf("can't get namespace of cluster-scoped resource: %+v", n)
		logrus.Fatalf("%+v", err)
	}
	if *n.Namespace == "" {
		return "default"
	}
	return *n.Namespace
}

func ParseResources(values []interface{}) (*KubeResources, error) {
	resources := &KubeResources{
		ClusterScoped: map[string]map[string]*Node{},
		Namespaced:    map[string]map[string]map[string]*Node{},
	}

	for _, v := range values {
		switch m := v.(type) {
		case map[string]interface{}:
			kind := m["kind"].(string)
			metadata := m["metadata"].(map[string]interface{})
			namespace := metadata["namespace"].(string)
			name := metadata["name"].(string)
			logrus.Debugf("kind: %s; name: %s; namespace: %s\n", m["kind"], metadata["name"], namespace)
			err := resources.Add(NewNode(kind, namespace, name))
			utils.DoOrDie(err)
		case nil:
			logrus.Warnf("skipping nil object")
		default:
			return nil, errors.Errorf("invalid type: %+v", v)
		}
	}
	return resources, nil
}
