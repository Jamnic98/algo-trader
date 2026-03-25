package strategies

import "fmt"

type ParamType string

const (
	ParamTypeNumber ParamType = "number"
	ParamTypeString ParamType = "string"
	ParamTypeBool   ParamType = "boolean"
	ParamTypeSelect ParamType = "select"
)

type ParamSchema struct {
	Key     string    `json:"key"`
	Label   string    `json:"label"`
	Type    ParamType `json:"type"`
	Default any       `json:"default"`
	Min     *float64  `json:"min,omitempty"`
	Max     *float64  `json:"max,omitempty"`
	Options []string  `json:"options,omitempty"`
}

type StrategySchema struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Params      []ParamSchema `json:"params"`
}

type Constructor func(paramsJSON []byte) (Strategy, error)

type registration struct {
	schema      StrategySchema
	constructor Constructor
}

var registry = map[string]registration{}

func Register(name string, schema StrategySchema, fn Constructor) {
	registry[name] = registration{schema: schema, constructor: fn}
}

func Build(name string, paramsJSON []byte) (Strategy, error) {
	r, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown strategy %q", name)
	}
	return r.constructor(paramsJSON)
}

func Schemas() []StrategySchema {
	schemas := make([]StrategySchema, 0, len(registry))
	for _, r := range registry {
		schemas = append(schemas, r.schema)
	}
	return schemas
}
