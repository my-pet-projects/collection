package gemini

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/genai"
)

func TestParseResult(t *testing.T) {
	t.Parallel()

	result, err := parseResult(`{
		"recognized": true,
		"beerName": "  Punk IPA ",
		"beerType": "IPA",
		"style": "India Pale Ale",
		"brewery": "BrewDog",
		"confidence": 0.94,
		"notes": ""
	}`)
	if err != nil {
		t.Fatalf("parseResult() error = %v", err)
	}
	if result.BeerName != "Punk IPA" {
		t.Fatalf("BeerName = %q, want Punk IPA", result.BeerName)
	}
	if result.Confidence != 0.94 {
		t.Fatalf("Confidence = %v, want 0.94", result.Confidence)
	}
}

func TestParseResultRejectsMalformedRecognition(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"invalid JSON":        `{`,
		"confidence too high": `{"recognized":false,"confidence":1.2}`,
		"missing beer name":   `{"recognized":true,"confidence":0.8}`,
	}
	for name, raw := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseResult(raw)
			if err == nil {
				t.Fatal("parseResult() error = nil, want error")
			}
		})
	}
}

func TestResponseTextRejectsIncompleteResponse(t *testing.T) {
	t.Parallel()

	response := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{
			FinishReason:  genai.FinishReasonMaxTokens,
			FinishMessage: "output limit reached",
		}},
	}

	_, err := responseText(response)
	if err == nil || !strings.Contains(err.Error(), "MAX_TOKENS") {
		t.Fatalf("responseText() error = %v, want MAX_TOKENS error", err)
	}
}

func TestResponseTextRejectsBlockedPrompt(t *testing.T) {
	t.Parallel()

	response := &genai.GenerateContentResponse{
		PromptFeedback: &genai.GenerateContentResponsePromptFeedback{
			BlockReason: genai.BlockedReasonSafety,
		},
	}

	_, err := responseText(response)
	if err == nil || !strings.Contains(err.Error(), "SAFETY") {
		t.Fatalf("responseText() error = %v, want SAFETY error", err)
	}
}

func TestRetryDelay(t *testing.T) {
	t.Parallel()

	details := []map[string]any{{
		"@type":      "type.googleapis.com/google.rpc.RetryInfo",
		"retryDelay": "13112s",
	}}

	if got := retryDelay(details); got != 13112*time.Second {
		t.Fatalf("retryDelay() = %s, want 13112s", got)
	}
}
