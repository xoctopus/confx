package exporter

import (
	"encoding/base64"

	"github.com/go-json-experiment/json/jsontext"
)

func TokenFor(v any) []jsontext.Token {
	switch x := v.(type) {
	case string:
		return []jsontext.Token{jsontext.String(x)}
	case int:
		return []jsontext.Token{jsontext.Int(int64(x))}
	case int8:
		return []jsontext.Token{jsontext.Int(int64(x))}
	case int16:
		return []jsontext.Token{jsontext.Int(int64(x))}
	case int32:
		return []jsontext.Token{jsontext.Int(int64(x))}
	case int64:
		return []jsontext.Token{jsontext.Int(x)}
	case uint:
		return []jsontext.Token{jsontext.Uint(uint64(x))}
	case uint8:
		return []jsontext.Token{jsontext.Uint(uint64(x))}
	case uint16:
		return []jsontext.Token{jsontext.Uint(uint64(x))}
	case uint32:
		return []jsontext.Token{jsontext.Uint(uint64(x))}
	case uint64:
		return []jsontext.Token{jsontext.Uint(x)}
	case float32:
		return []jsontext.Token{jsontext.Float(float64(x))}
	case float64:
		return []jsontext.Token{jsontext.Float(x)}
	case bool:
		return []jsontext.Token{jsontext.Bool(x)}
	case map[string]any:
		tokens := []jsontext.Token{jsontext.BeginObject}
		for k := range x {
			tokens = append(tokens, jsontext.String(k))
			tokens = append(tokens, TokenFor(x[k])...)
		}
		tokens = append(tokens, jsontext.EndObject)
		return tokens
	case []any:
		tokens := []jsontext.Token{jsontext.BeginArray}
		for i := range x {
			tokens = append(tokens, TokenFor(x[i])...)
		}
		tokens = append(tokens, jsontext.EndArray)
		return tokens
	case []byte:
		return []jsontext.Token{jsontext.String(base64.StdEncoding.EncodeToString(x))}
	}
	return nil
}
