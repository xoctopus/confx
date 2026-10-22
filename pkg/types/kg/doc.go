// Package kg defines a structured naming convention for cache keys.
//
// Keys use four segments joined by [SegmentSeparator] (":"):
//
//	{creator}:{audience}:{domain}:{biz}
//
// An optional tenant or site namespace may be prepended (not part of creator):
//
//	{prefix}:{creator}:{audience}:{domain}:{biz}
//
// # Segments
//
//   - prefix: Tenant or site id via [WithPrefix]; separate from creator so
//     service identity and [KeyGen.SharedKeyFrom] stay stable across sites.
//   - creator: Service that owns the key, from [KeyGen.Init].
//   - audience: [KeyGen.InstanceID], [AudiencePeer], a named service, or [AudienceAny].
//   - domain: Caller-defined business area (for example SESSION).
//   - biz: Resource id within domain; [KeyType] is ~uint64 or ~string.
//
// # Segment rules
//
// [SanitizeSegment] trims and replaces whitespace, [SegmentSeparator], and hyphen
// with underscore. [NormalizeSegment] also uppercases (creator, audience, prefix).
// After sanitization, domain and biz must be non-empty or [KeyGen] panics.
// [NormalizeCreator] and [NormalizeAudience] reject empty results and reserved
// words [AudiencePeer] and [AudienceAny].
//
// String biz values are sanitized like domain (hyphens become underscores).
// Prefer uint64 primary keys for numeric ids; signed integers are not [KeyType].
//
// # KeyGen
//
// One [KeyGen] per process; [KeyGen.Init] runs once. Use [KeyGen.Key] for
// process-local cache, [KeyGen.SharedKey] for same-service sharing,
// [KeyGen.ShareTo] for a named consumer, [KeyGen.GlobalKey] for cross-service
// keys. [KeyGen.SharedKeyFrom], [KeyGen.ShareFrom], and [KeyGen.GlobalKeyFrom]
// address keys built by other creators (with the same prefix when configured).
package kg
