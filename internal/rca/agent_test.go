package rca

import "testing"

// Regression: Groq's gpt-oss models sometimes declare their final answer by
// calling a nonexistent "json" tool; the provider rejects the generation with
// 400 but embeds it under failed_generation. The salvage path must recover it.
func TestSalvageFinalFromError(t *testing.T) {
	errStr := `llm returned status 400: {"error":{"message":"Tool call validation failed: attempted to call tool 'json' which was not in request.tools","type":"invalid_request_error","code":"tool_use_failed","failed_generation":"{\"name\": \"json\", \"arguments\": {\n  \"summary\": \"Memory limit exceeded\",\n  \"root_cause\": \"Container was OOMKilled after deployment\",\n  \"confidence\": 0.75,\n  \"suggested_actions\": \"restart\",\n  \"evidence_logs\": [],\n  \"evidence_events\": [\"BackOff\"]\n}}"}}`

	final, ok := salvageFinalFromError(&errWrap{errStr})
	if !ok {
		t.Fatal("expected salvage to succeed on tool-use-failed error with embedded final")
	}
	if final.rootCause != "Container was OOMKilled after deployment" {
		t.Fatalf("wrong root cause salvaged: %q", final.rootCause)
	}
	if final.confidence != 0.75 {
		t.Fatalf("wrong confidence: %v", final.confidence)
	}
}

// Non-recovery errors must return false so the fallback chain proceeds.
func TestSalvageFinalFromErrorRejected(t *testing.T) {
	if _, ok := salvageFinalFromError(&errWrap{`llm returned status 401: unauthorized`}); ok {
		t.Fatal("salvage must not fire on non-tool-use errors")
	}
}

type errWrap struct{ s string }

func (e *errWrap) Error() string { return e.s }
