package strategies

import (
	"encoding/json"
	"fmt"
	"os"
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

// Definitions returns the parsed shared/strategies.json (loaded once).
func Definitions() *SchemaFile {
	defsOnce.Do(func() {
		path := schemaPath()
		data, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("strategies: cannot load schema %q: %v", path, err))
		}
		var sf SchemaFile
		if err := json.Unmarshal(data, &sf); err != nil {
			panic(fmt.Sprintf("strategies: invalid schema %q: %v", path, err))
		}
		if sf.Composite.DefaultThreshold <= 0 {
			sf.Composite.DefaultThreshold = 0.6
		}
		defs = &sf
	})
	return defs
}

// schemaPath resolves the shared schema file: STRATEGIES_SCHEMA_PATH env, then
// candidate relative paths (works from repo root, trader-core/, any package dir,
// and inside containers where shared/ is copied next to the app).
func schemaPath() string {
	if p := os.Getenv("STRATEGIES_SCHEMA_PATH"); p != "" {
		return p
	}
	candidates := []string{
		"shared/strategies.json",
		"../shared/strategies.json",
		"../../shared/strategies.json",
		"../../../shared/strategies.json",
		"../../../../shared/strategies.json",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
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

// init registers every primitive from the shared schema, keyed by template.
func init() {
	for _, def := range Definitions().Primitives {
		d := def
		Register(d.Name, StrategySchema{
			Name:        d.Name,
			DisplayName: d.DisplayName,
			Params:      d.Params,
		}, makeConstructor(d))
	}
}

// withDefaults fills missing params from the schema defaults.
func withDefaults(def PrimitiveDef, paramsJSON []byte) []byte {
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

// makeConstructor maps a schema template name to the Go implementation.
// Adding a new template means implementing it here AND in research/app/strategies;
// the shared test vectors in strategies.json pin the behavior on both sides.
func makeConstructor(def PrimitiveDef) Constructor {
	return func(paramsJSON []byte) (Strategy, error) {
		params := withDefaults(def, paramsJSON)
		switch def.Template {
		case "always_buy":
			return NewSimpleDCAStrategy(SimpleDCAParams{}), nil
		case "prev_close":
			return NewSimpleStrategy(SimpleParams{}), nil
		case "dca":
			var p SmartDCAParams
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("parsing %s params: %w", def.Name, err)
			}
			return NewSmartDCAStrategy(p), nil
		default:
			return nil, fmt.Errorf("unknown template %q for strategy %q", def.Template, def.Name)
		}
	}
}
