package bot

import (
	"encoding/json"
	"testing"
)

func TestMergeParams(t *testing.T) {
	// empty overrides -> base returned as-is
	base := json.RawMessage(`{"a":1}`)
	got, err := mergeParams(base, nil)
	if err != nil {
		t.Fatalf("mergeParams: %v", err)
	}
	if string(got) != `{"a":1}` {
		t.Fatalf("got %s", got)
	}

	// overrides win over base keys, new keys added
	got, err = mergeParams(base, json.RawMessage(`{"b":2,"a":9}`))
	if err != nil {
		t.Fatalf("mergeParams: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["a"] != float64(9) || m["b"] != float64(2) {
		t.Fatalf("merged = %v", m)
	}

	// empty base + overrides
	got, err = mergeParams(nil, json.RawMessage(`{"x":"y"}`))
	if err != nil {
		t.Fatalf("mergeParams: %v", err)
	}
	if string(got) != `{"x":"y"}` {
		t.Fatalf("got %s", got)
	}

	// invalid override JSON -> error
	if _, err := mergeParams(base, json.RawMessage(`{bad`)); err == nil {
		t.Fatal("expected error for invalid overrides")
	}
}

func TestBuildBotConfigDefaultsAccount(t *testing.T) {
	cfg, err := BuildBotConfig("id1", CreateBotData{Mode: BotModePaper, StrategySlug: "simple", Interval: "1h"})
	if err != nil {
		t.Fatalf("BuildBotConfig: %v", err)
	}
	if cfg.AccountName != "default" {
		t.Fatalf("AccountName = %q, want default", cfg.AccountName)
	}

	cfg2, err := BuildBotConfig("id2", CreateBotData{Mode: BotModeLive, Account: "live", StrategySlug: "simple", Interval: "1h"})
	if err != nil {
		t.Fatalf("BuildBotConfig live: %v", err)
	}
	if cfg2.AccountName != "live" || cfg2.Mode != BotModeLive {
		t.Fatalf("cfg2 = %+v", cfg2)
	}
}
