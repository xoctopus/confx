package kg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/xoctopus/x/misc/must"
)

// Key format (four segments):
//
//	{creator}:{audience}:{domain}:{biz}
//
//   - creator:  the owning service/process identity, set by [KeyGen.Init].
//   - audience: who may read the key (instance ULID, PEER, a named service, or ANY).
//   - domain:   business domain chosen by the caller.
//   - biz:      domain-specific business identifier.
//
// Namespace prefixes (e.g. site_id) are intentionally excluded;
// callers who need tenant isolation should prepend their own prefix
// via [Option] or wrap [KeyGen].

// Audience reserved words.
const (
	AudiencePeer = "PEER"
	AudienceAny  = "ANY"
)

// KeyCode constrains the biz-key segment to common scalar types.
type KeyCode interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 | ~string
}

// Option configures a [KeyGen] at Init time.
type Option func(*KeyGen)

// WithPrefix prepends a static namespace segment before the four-segment key,
// producing {prefix}:{creator}:{audience}:{domain}:{biz}.
func WithPrefix(prefix string) Option {
	return func(g *KeyGen) { g.prefix = strings.TrimSpace(prefix) }
}

// KeyGen generates structured cache keys using a four-segment convention.
// It must be initialized exactly once via [KeyGen.Init] before any key
// generation method is called.
type KeyGen struct {
	once    sync.Once
	prefix  string
	creator string
	inst    string
}

// Init locks the creator identity and generates a per-process ULID instance ID.
// Subsequent calls are no-ops (the first result is retained).
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

// Creator returns the normalized creator identity.
func (g *KeyGen) Creator() string { return g.creator }

// Prefix returns the optional namespace prefix.
func (g *KeyGen) Prefix() string { return g.prefix }

// InstanceID returns the per-process instance audience (ULID).
func (g *KeyGen) InstanceID() string { return g.inst }

// Key returns a process-exclusive key: {creator}:{ulid}:{domain}:{biz}.
func (g *KeyGen) Key(domain string, k any) string {
	return g.join(g.creator, g.inst, domain, k)
}

// SharedKey returns a peer-shared key: {creator}:PEER:{domain}:{biz}.
func (g *KeyGen) SharedKey(domain string, k any) string {
	return g.join(g.creator, AudiencePeer, domain, k)
}

// ShareTo returns a key shared with a target identity: {creator}:{audience}:{domain}:{biz}.
func (g *KeyGen) ShareTo(audienceIdentity, domain string, k any) string {
	return g.join(g.creator, NormalizeAudience(audienceIdentity), domain, k)
}

// GlobalKey returns a cross-service key: {creator}:ANY:{domain}:{biz}.
func (g *KeyGen) GlobalKey(domain string, k any) string {
	return g.join(g.creator, AudienceAny, domain, k)
}

// SharedKeyFrom reads a PEER key created by another creator identity.
func (g *KeyGen) SharedKeyFrom(creatorIdentity, domain string, k any) string {
	return g.join(NormalizeCreator(creatorIdentity), AudiencePeer, domain, k)
}

// ShareFrom reads a key shared by creator identity to audience identity.
func (g *KeyGen) ShareFrom(creatorIdentity, audienceIdentity, domain string, k any) string {
	return g.join(NormalizeCreator(creatorIdentity), NormalizeAudience(audienceIdentity), domain, k)
}

// GlobalKeyFrom reads a cross-service key created by another creator identity.
func (g *KeyGen) GlobalKeyFrom(creatorIdentity, domain string, k any) string {
	return g.join(NormalizeCreator(creatorIdentity), AudienceAny, domain, k)
}

func (g *KeyGen) join(creator, audience, domain string, k any) string {
	if g.prefix != "" {
		return fmt.Sprintf("%s:%s:%s:%s:%v", g.prefix, creator, audience, domain, k)
	}
	return fmt.Sprintf("%s:%s:%s:%v", creator, audience, domain, k)
}

// NormalizeSegment normalises a key segment: TrimSpace, ToUpper, '-' to '_'.
func NormalizeSegment(name string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), "-", "_"))
}

// NormalizeCreator normalizes creator identity for the creator segment;
// panics on empty or reserved words.
func NormalizeCreator(name string) string {
	c := NormalizeSegment(name)
	must.BeTrueF(c != "", "kgen: creator is empty")
	must.BeTrueF(c != AudiencePeer && c != AudienceAny, "kgen: creator is reserved %q", c)
	return c
}

// NormalizeAudience normalizes audience identity for the audience segment;
// panics on empty or reserved words.
func NormalizeAudience(name string) string {
	a := NormalizeSegment(name)
	must.BeTrueF(a != "", "kgen: audience is empty")
	must.BeTrueF(a != AudiencePeer && a != AudienceAny, "kgen: audience is reserved %q", a)
	return a
}
