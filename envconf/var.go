package envconf

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func NewVar(name, value string) *Var {
	return &Var{Name: name, Value: value}
}

type Var struct {
	Name  string
	Value string
	Mask  string

	Options any
}

func (v *Var) GroupName(prefix string) string {
	return prefix + "__" + v.Name
}

func (v *Var) ParseOption(flag map[string]struct{}) {
	// todo to parse options from flag
}

func NewGroup(name string) *Group {
	return &Group{
		Name:   name,
		Values: make(map[string]*Var),
	}
}

type Group struct {
	Name   string
	Values map[string]*Var
}

func (g *Group) MapEntries(k string) (entries []string) {
	for _, v := range g.Values {
		if !strings.HasPrefix(v.Name, k) {
			continue
		}

		if entry := strings.TrimPrefix(v.Name, k+"_"); len(entry) > 0 {
			entries = append(entries, entry)
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i] < entries[j]
	})
	return
}

func (g *Group) SliceLength(k string) int {
	size := -1

	for _, v := range g.Values {
		if !strings.HasPrefix(v.Name, k) {
			continue
		}

		suffix := strings.TrimPrefix(v.Name, k+"_")
		idx, err := strconv.ParseInt(strings.Split(suffix, "_")[0], 10, 64)
		if err == nil && int(idx) > size {
			size = int(idx)
		}
	}

	return size + 1
}

func (g *Group) Len() int {
	return len(g.Values)
}

func (g *Group) Get(name string) *Var {
	return g.Values[name]
}

func (g *Group) Add(v *Var) {
	g.Values[v.Name] = v
}

func (g *Group) Del(name string) {
	delete(g.Values, name)
}

func (g *Group) Reset() {
	g.Values = make(map[string]*Var)
}

func (g *Group) Print() {
	for _, v := range g.Values {
		fmt.Printf("%s: %s\n", v.GroupName(g.Name), v.Value)
	}
}
