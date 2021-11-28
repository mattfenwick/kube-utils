package graph

import (
	"fmt"
	"strings"
)

type Graph struct {
	Name      string
	Label     string
	Nodes     map[string][]string
	Edges     map[string]map[string]bool
	Subgraphs map[string]*Graph
}

func NewGraph(name string, label string) *Graph {
	return &Graph{
		Name:      name,
		Label:     label,
		Nodes:     map[string][]string{},
		Edges:     map[string]map[string]bool{},
		Subgraphs: map[string]*Graph{},
	}
}

// AddNode adds a node
// example: `  "%s" [color="%s",penwidth=2,style="dashed"];`
func (g *Graph) AddNode(node string, config ...string) {
	if _, ok := g.Edges[node]; !ok {
		g.Nodes[node] = config
		g.Edges[node] = map[string]bool{}
	}
}

func (g *Graph) AddEdge(from string, to string) {
	if _, ok := g.Nodes[from]; !ok {
		g.AddNode(from)
	}
	if _, ok := g.Nodes[to]; !ok {
		g.AddNode(to)
	}
	g.Edges[from][to] = true
}

func (g *Graph) AddSubgraph(sub *Graph) {
	g.Subgraphs[sub.Name] = sub
}

func (g *Graph) RenderDotBody(indent string) []string {
	lines := []string{
		fmt.Sprintf(`%s  label="%s";`, indent, g.Label),
	}

	for node, config := range g.Nodes {
		lines = append(lines, fmt.Sprintf(`%s  "%s" [%s];`, indent, node, strings.Join(config, ", ")))
	}
	lines = append(lines, "")

	for from, tos := range g.Edges {
		for to := range tos {
			lines = append(lines, fmt.Sprintf(`%s  "%s" -> "%s" [color=red,penwidth=5,style="dashed"];`, indent, from, to))
		}
	}
	lines = append(lines, "")

	for _, sub := range g.Subgraphs {
		lines = append(lines, fmt.Sprintf(`%s  subgraph "cluster_%s" {`, indent, sub.Name))
		lines = append(lines, sub.RenderDotBody(indent+"  ")...)
		lines = append(lines, indent+"  }")
	}
	return lines
}

func (g *Graph) RenderAsDot() string {
	lines := []string{fmt.Sprintf(`digraph "%s" {`, g.Name)}
	lines = append(lines, g.RenderDotBody("")...)
	return strings.Join(append(lines, "}"), "\n")
}
