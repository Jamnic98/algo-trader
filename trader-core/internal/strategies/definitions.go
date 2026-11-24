package strategies

import (
	"encoding/json"
	"fmt"
	"sync"
)

// TestCase is a shared behavior vector: feed candles (closes) in order and
// expect these signals back. Both the Go and Python engines run these.
type TestCase struct {
	Candles []float64 `json:"candles"`
	Expect  []string  `json:"expect"`
}

type TestVector struct {
	FeeRate string     `json:"fee_rate"`
	Cases   []TestCase `json:"cases"`
}

// PrimitiveDef is one strategy definition from shared/strategies.json.
type PrimitiveDef struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Template    string        `json:"template"`
	Params      []ParamSchema `json:"params"`
	Test        TestVector    `json:"test"`
}

type CompositeDef struct {
	Template         string  `json:"template"`
	DefaultThreshold float64 `json:"default_threshold"`
}

type SchemaFile struct {
	Version    int            `json:"version"`
	Composite  CompositeDef   `json:"composite"`
	Primitives []PrimitiveDef `json:"primitives"`
}

var (
	defsOnce sync.Once
	defs     *SchemaFile
)

// Definitions returns the parsed schema. The JSON is embedded in the binary
// via gen_schema.go (regenerate with scripts/generate_strategies.py after
// editing shared/strategies.json).
func Definitions() *SchemaFile {
	defsOnce.Do(func() {
		var sf SchemaFile
		if err := json.Unmarshal([]byte(generatedSchemaJSON), &sf); err != nil {
			panic(fmt.Sprintf("strategies: invalid embedded schema: %v", err))
		}
		if sf.Composite.DefaultThreshold <= 0 {
			sf.Composite.DefaultThreshold = 0.6
		}
		defs = &sf
	})
	return defs
}

// FindPrimitive returns the definition for a strategy slug, or nil.
func FindPrimitive(name string) *PrimitiveDef {
	for i := range Definitions().Primitives {
		if Definitions().Primitives[i].Name == name {
			return &Definitions().Primitives[i]
		}
	}
	return nil
}

// withDefaults fills missing params from the schema defaults.
func withDefaults(name string, paramsJSON []byte) []byte {
	def := FindPrimitive(name)
	if def == nil {
		return paramsJSON
	}
	m := map[string]any{}
	if len(paramsJSON) > 0 {
		if err := json.Unmarshal(paramsJSON, &m); err != nil {
			return paramsJSON
		}
	}
	for _, p := range def.Params {
		if _, ok := m[p.Key]; !ok && p.Default != nil {
			m[p.Key] = p.Default
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		return paramsJSON
	}
	return out
}

// constructorFor maps a schema template to the Go implementation. New
// templates must be implemented here AND in research/app/strategies; the
// shared test vectors in shared/strategies.json pin behavior on both sides.
func constructorFor(name, template string) Constructor {
	return func(paramsJSON []byte) (Strategy, error) {
		params := withDefaults(name, paramsJSON)
		switch template {
		case "always_buy":
			return NewSimpleDCAStrategy(SimpleDCAParams{}), nil
		case "prev_close":
			return NewSimpleStrategy(SimpleParams{}), nil
		case "dca":
			var p SmartDCAParams
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("parsing %s params: %w", name, err)
			}
			return NewSmartDCAStrategy(p), nil
		default:
			return nil, fmt.Errorf("unknown template %q for strategy %q", template, name)
		}
	}
}

func ptrFloat(f float64) *float64 { return &f }
