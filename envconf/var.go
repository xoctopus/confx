package envconf

import (
	"bytes"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

func NewVar(key, val string) *Var {
	return &Var{
		key: key,
		val: val,
	}
}

type Var struct {
	key  string
	val  string
	mask string

	optional bool
}

func (v *Var) Key() string {
	return v.key
}

func (v *Var) Value() string {
	return v.val
}

func NewGroup(name string) *Group {
	if len(name) == 0 || !alphabet(name) {
		panic("invalid group name: " + name)
	}
	return &Group{
		name: name,
		vars: make(map[string]*Var),
	}
}

func ParseGroupFromEnv(prefix string) *Group {
	g := NewGroup(prefix)
	for _, environ := range os.Environ() {
		kv := strings.SplitN(environ, "=", 2)
		if len(kv) == 2 {
			if strings.HasPrefix(kv[0], prefix) {
				g.Add(&Var{
					key: strings.TrimPrefix(kv[0], prefix+"__"),
					val: kv[1],
				})
			}
		}
	}
	return g
}

type Group struct {
	name string
	vars map[string]*Var
}

func (g *Group) MapEntries(k string) []string {
	keys := make(map[string]struct{})
	for _, v := range g.vars {
		if !strings.HasPrefix(v.key, k) {
			continue
		}
		// map key must be quoted with `_`
		entry := strings.Trim(strings.TrimPrefix(v.key, k), "_")
		index := strings.Index(entry, "_")
		if index > 0 {
			entry = strings.Trim(entry[:index], "_")
		}
		if len(entry) > 0 {
			keys[entry] = struct{}{}
		}
	}
	return maps.Keys(keys)
}

func (g *Group) SliceLength(k string) int {
	size := -1

	for _, v := range g.vars {
		if !strings.HasPrefix(v.key, k) {
			continue
		}

		entry := strings.Trim(strings.TrimPrefix(v.key, k), "_")
		index := strings.Index(entry, "_")
		if index > 0 {
			entry = strings.Trim(entry[:index], "_")
		}
		if len(entry) > 0 {
			idx, err := strconv.ParseInt(entry, 10, 64)
			if err == nil && int(idx) > size {
				size = int(idx)
			}
		}
	}

	return size + 1
}

func (g *Group) Len() int {
	return len(g.vars)
}

func (g *Group) Get(key string) *Var {
	return g.vars[key]
}

func (g *Group) Key(key string) string {
	return g.name + "__" + key
}

func (g *Group) Add(v *Var) bool {
	_, ok := g.vars[v.key]
	g.vars[v.key] = v
	return ok
}

func (g *Group) Bytes() []byte {
	return g.dotenv(nil)
}

func (g *Group) MaskBytes() []byte {
	return g.dotenv(func(v *Var) string { return v.mask })
}

func (g *Group) dotenv(valuer func(*Var) string) []byte {
	values := make(map[string]string)
	for _, v := range g.vars {
		val := v.val
		if valuer != nil {
			val = valuer(v)
		}
		values[g.Key(v.key)] = val
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
