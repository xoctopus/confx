package envconf

import (
	"bytes"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

type VarOptions struct {
	Optional bool
}

func NewVar(name, value string) *Var {
	return &Var{Name: name, Value: value}
}

type Var struct {
	Name  string
	Value string
	Mask  string

	VarOptions
}

func (v *Var) GroupName(prefix string) string {
	return prefix + "__" + v.Name
}

func (v *Var) ParseOption(flag map[string]struct{}) {
	if _, ok := flag["optional"]; ok {
		v.Optional = true
	}
}

func NewGroup(name string) *Group {
	return &Group{
		Name:   name,
		Values: make(map[string]*Var),
	}
}

func ParseGroupFromEnv(prefix string) *Group {
	g := NewGroup(prefix)
	for _, environ := range os.Environ() {
		kv := strings.SplitN(environ, "=", 2)
		if len(kv) == 2 {
			if strings.HasPrefix(kv[0], prefix) {
				g.Add(&Var{
					Name:  strings.TrimPrefix(kv[0], prefix+"__"),
					Value: kv[1],
				})
			}
		}
	}
	return g
}

type Group struct {
	Name   string
	Values map[string]*Var
}

func (g *Group) MapEntries(k string) []string {
	keys := make(map[string]struct{})
	for _, v := range g.Values {
		if !strings.HasPrefix(v.Name, k+"_") {
			continue
		}

		if entry := strings.TrimPrefix(v.Name, k+"_"); len(entry) > 0 {
			if index := strings.Index(entry, "_"); index > 0 {
				entry = entry[:index]
			}
			keys[entry] = struct{}{}
		}
	}
	entries := maps.Keys(keys)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i] < entries[j]
	})
	return entries
}

func (g *Group) SliceLength(k string) int {
	size := -1

	for _, v := range g.Values {
		if !strings.HasPrefix(v.Name, k+"_") {
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

func (g *Group) Bytes() []byte {
	return g.DotEnv(nil)
}

func (g *Group) MaskBytes() []byte {
	return g.DotEnv(func(v *Var) string {
		if v.Mask != "" {
			return v.Mask
		}
		return v.Value
	})
}

func (g *Group) DotEnv(valuer func(*Var) string) []byte {
	values := make(map[string]string)
	for _, v := range g.Values {
		if valuer != nil {
			values[v.GroupName(g.Name)] = valuer(v)
		} else {
			values[v.GroupName(g.Name)] = v.Value
		}
	}
	return DotEnv(values)
}

func DotEnv(values map[string]string) []byte {
	buf := bytes.NewBuffer(nil)

	keys := maps.Keys(values)
	sort.Strings(keys)

	for _, key := range keys {
		buf.WriteString(key)
		buf.WriteRune('=')
		buf.WriteString(values[key])
		buf.WriteRune('\n')
	}

	return buf.Bytes()
}
