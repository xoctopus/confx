package kg

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/oklog/ulid/v2"
	"github.com/xoctopus/x/misc/must"
)

// SegmentSeparator joins key segments. See package documentation for format.
const SegmentSeparator = ":"

// AudiencePeer is the audience segment for keys shared by all instances of
// the same creator. Used by [KeyGen.SharedKey].
const AudiencePeer = "PEER"

// AudienceAny is the audience segment for keys visible across services.
// Used by [KeyGen.GlobalKey].
const AudienceAny = "ANY"

// KeyType is the resource id for the biz segment (~uint64 or ~string).
type KeyType interface {
	~uint64 | ~string
}

// Option configures [KeyGen] on the first [KeyGen.Init] only.
type Option func(*KeyGen)

// WithPrefix sets an optional tenant or site namespace prepended to every key.
// Sanitized and uppercased; empty after normalization omits the prefix.
func WithPrefix(prefix string) Option {
	return func(g *KeyGen) { g.prefix = NormalizeSegment(prefix) }
}

// WithInstanceID sets the audience segment for [KeyGen.Key] only.
// Sanitized but not uppercased. Empty defers to a ULID at Init.
func WithInstanceID(id string) Option {
	return func(g *KeyGen) { g.inst = SanitizeSegment(id) }
}

// KeyGen generates cache keys after [KeyGen.Init].
type KeyGen struct {
	once    sync.Once
	prefix  string
	creator string
	inst    string
}

// Init sets creator, applies options, and allocates [KeyGen.InstanceID] when
// [WithInstanceID] did not supply one. Later calls are ignored.
func (g *KeyGen) Init(creator string, opts ...Option) {
	g.once.Do(func() {
		for _, o := range opts {
			o(g)
		}
		g.creator = NormalizeCreator(creator)
		if g.inst == "" {
			g.inst = ulid.Make().String()
		}
	})
}

// Creator returns the normalized creator from Init.
func (g *KeyGen) Creator() string { return g.creator }

// Prefix returns the optional namespace from [WithPrefix].
func (g *KeyGen) Prefix() string { return g.prefix }

// InstanceID returns the audience used by [KeyGen.Key].
func (g *KeyGen) InstanceID() string { return g.inst }

// Key formats creator:instanceID:domain:biz (plus prefix when set).
func (g *KeyGen) Key[K KeyType](domain string, k K) string {
	return g.gen(g.creator, g.inst, domain, k)
}

// SharedKey formats creator:PEER:domain:biz.
func (g *KeyGen) SharedKey[K KeyType](domain string, k K) string {
	return g.gen(g.creator, AudiencePeer, domain, k)
}

// ShareTo formats creator:audience:domain:biz for a named service identity.
func (g *KeyGen) ShareTo[K KeyType](audienceIdentity, domain string, k K) string {
	return g.gen(g.creator, NormalizeAudience(audienceIdentity), domain, k)
}

// GlobalKey formats creator:ANY:domain:biz.
func (g *KeyGen) GlobalKey[K KeyType](domain string, k K) string {
	return g.gen(g.creator, AudienceAny, domain, k)
}

// SharedKeyFrom formats creator:PEER:domain:biz for another creator.
func (g *KeyGen) SharedKeyFrom[K KeyType](creator, domain string, k K) string {
	return g.gen(NormalizeCreator(creator), AudiencePeer, domain, k)
}

// ShareFrom formats creator:audience:domain:biz for another creator.
func (g *KeyGen) ShareFrom[K KeyType](creator, audience, domain string, k K) string {
	return g.gen(NormalizeCreator(creator), NormalizeAudience(audience), domain, k)
}

// GlobalKeyFrom formats creator:ANY:domain:biz for another creator.
func (g *KeyGen) GlobalKeyFrom[K KeyType](creator, domain string, k K) string {
	return g.gen(NormalizeCreator(creator), AudienceAny, domain, k)
}

func (g *KeyGen) gen[K KeyType](creator, audience, domain string, k K) string {
	biz := fmt.Sprintf("%v", k)
	domain = SanitizeSegment(domain)
	biz = SanitizeSegment(biz)
	must.BeTrueF(len(domain) > 0, "kgen: domain segment is empty")
	must.BeTrueF(len(biz) > 0, "kgen: biz segment is empty")

	segs := []string{creator, audience, domain, biz}
	if g.prefix != "" {
		segs = append([]string{g.prefix}, segs...)
	}
	return strings.Join(segs, SegmentSeparator)
}

// SanitizeSegment trims and maps whitespace, [SegmentSeparator], and hyphen to underscore.
func SanitizeSegment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, SegmentSeparator, "_")
	s = strings.ReplaceAll(s, "-", "_")
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			b.WriteByte('_')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// NormalizeSegment sanitizes and uppercases a segment (creator, audience, prefix).
func NormalizeSegment(name string) string {
	return strings.ToUpper(SanitizeSegment(name))
}

// NormalizeCreator prepares the creator segment.
func NormalizeCreator(name string) string {
	c := NormalizeSegment(name)
	must.BeTrueF(len(c) > 0, "kgen: creator segment is empty")
	must.BeTrueF(c != AudiencePeer && c != AudienceAny, "kgen: creator is reserved %q", c)
	return c
}

// NormalizeAudience prepares a named service audience segment.
func NormalizeAudience(name string) string {
	a := NormalizeSegment(name)
	must.BeTrueF(len(a) > 0, "kgen: audience segment is empty")
	must.BeTrueF(a != AudiencePeer && a != AudienceAny, "kgen: audience is reserved %q", a)
	return a
}
