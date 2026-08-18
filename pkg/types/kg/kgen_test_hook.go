package kg

import "sync"

// reset resets internal state for tests only
func reset(g *KeyGen) {
	g.once = sync.Once{}
	g.prefix = ""
	g.creator = ""
	g.inst = ""
}
