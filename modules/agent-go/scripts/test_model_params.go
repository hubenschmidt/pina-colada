package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var openAIModels = []string{
	"gpt-5.2",
	"gpt-5.1",
	"gpt-5",
	"gpt-5-mini",
	"gpt-5-nano",
	"gpt-4.1",
	"gpt-4.1-mini",
	"gpt-4o",
	"gpt-4o-mini",
	"o3",
	"o4-mini",
}

var anthropicModels = []string{
	"claude-sonnet-4-5-20250514",
	"claude-opus-4-5-20250514",
	"claude-haiku-4-5-20251001",
}

// All ModelSettings params from OpenAI Agents SDK
var openAIParams = []struct {
	name  string
	value any
}{
	{"temperature", 0.7},
	{"top_p", 0.9},
	{"frequency_penalty", 0.5},
	{"presence_penalty", 0.5},
	{"max_tokens", 100},
	{"parallel_tool_calls", true},
	{"truncation", "auto"},
	{"store", true},
}

// Anthropic params
var anthropicParams = []struct {
	name  string
	value any
}{
	{"temperature", 0.7},
	{"top_p", 0.9},
	{"max_tokens", 100},
}

type ParamResult struct {
	Supported bool
	Error     string
}

type ModelResult struct {
	Model    string
	Provider string
	Params   map[string]ParamResult
}

func main() {
	// Load .env from parent directory (agent-go/)
	_ = godotenv.Load("../.env")

	openAIKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	if openAIKey == "" {
		fmt.Println("OPENAI_API_KEY not set")
		return
	}
	if anthropicKey == "" {
		fmt.Println("ANTHROPIC_API_KEY not set")
		return
	}

	fmt.Println("Testing model parameter support (concurrent)...")
	fmt.Println("================================================")
	fmt.Println()

	var wg sync.WaitGroup
	resultsChan := make(chan ModelResult, len(openAIModels)+len(anthropicModels))

	// Launch OpenAI model tests concurrently
	for _, model := range openAIModels {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			result := testOpenAIModel(m, openAIKey)
			resultsChan <- result
			fmt.Printf("✓ Tested %s\n", m)
		}(model)
	}

	// Launch Anthropic model tests concurrently
	for _, model := range anthropicModels {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			result := testAnthropicModel(m, anthropicKey)
			resultsChan <- result
			fmt.Printf("✓ Tested %s\n", m)
		}(model)
	}

	// Wait for all tests to complete
	wg.Wait()
	close(resultsChan)

	// Collect results
	var results []ModelResult
	for r := range resultsChan {
		results = append(results, r)
	}

	// Sort results by provider then model for consistent output
	sortResults(results)

	fmt.Println()

	// Print summary table
	printSummary(results)

	// Print Go code for unsupported params
	printGoCode(results)
}

func sortResults(results []ModelResult) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Provider != results[j].Provider {
			return results[i].Provider < results[j].Provider
		}
		return results[i].Model < results[j].Model
	})
}

func printSummary(results []ModelResult) {
	fmt.Println("==================================")
	fmt.Println("SUMMARY TABLE")
	fmt.Println("==================================")
	fmt.Println()

	// Collect all param names
	allParams := []string{"temperature", "top_p", "frequency_penalty", "presence_penalty", "max_tokens", "parallel_tool_calls", "truncation", "store"}

	// Header
	fmt.Printf("%-28s", "Model")
	for _, p := range allParams {
		fmt.Printf(" | %s", centerPad(abbrev(p), 8))
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 28+len(allParams)*11))

	// Rows
	for _, r := range results {
		fmt.Printf("%-28s", r.Model)
		for _, p := range allParams {
			fmt.Printf(" | %s", formatParamResult(r.Params[p]))
		}
		fmt.Println()
	}
}

func centerPad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func formatParamResult(res ParamResult) string {
	if res == (ParamResult{}) {
		return "  n/a   "
	}
	if !res.Supported {
		return "   ❌   "
	}
	return "   ✅   "
}

func abbrev(s string) string {
	abbrevs := map[string]string{
		"temperature":         "temp",
		"top_p":               "top_p",
		"frequency_penalty":   "freq_pen",
		"presence_penalty":    "pres_pen",
		"max_tokens":          "max_tok",
		"parallel_tool_calls": "parallel",
		"truncation":          "trunc",
		"store":               "store",
	}
	if a, ok := abbrevs[s]; ok {
		return a
	}
	return s[:8]
}

func printGoCode(results []ModelResult) {
	fmt.Println()
	fmt.Println("==================================")
	fmt.Println("GO CODE - Unsupported params map")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("var unsupportedModelParams = map[string][]string{")
	for _, r := range results {
		var unsupported []string
		for param, res := range r.Params {
			if !res.Supported {
				unsupported = append(unsupported, fmt.Sprintf("%q", param))
			}
		}
		if len(unsupported) > 0 {
			fmt.Printf("\t%q: {%s},\n", r.Model, strings.Join(unsupported, ", "))
		}
	}
	fmt.Println("}")
}

func testOpenAIModel(model, apiKey string) ModelResult {
	result := ModelResult{
		Model:    model,
		Provider: "openai",
		Params:   make(map[string]ParamResult),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range openAIParams {
		wg.Add(1)
		go func(param string, value any) {
			defer wg.Done()
			errMsg := testOpenAIWithParam(model, apiKey, param, value)
			supported := errMsg == "" || !strings.Contains(errMsg, "Unsupported parameter")

			mu.Lock()
			result.Params[param] = ParamResult{Supported: supported, Error: errMsg}
			mu.Unlock()
		}(p.name, p.value)
	}

	wg.Wait()
	return result
}

func testOpenAIWithParam(model, apiKey, param string, value any) string {
	payload := map[string]any{
		"model": model,
		"input": "Say hi",
	}
	payload[param] = value

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return ""
	}

	respBody, _ := io.ReadAll(resp.Body)
	return string(respBody)
}

func testAnthropicModel(model, apiKey string) ModelResult {
	result := ModelResult{
		Model:    model,
		Provider: "anthropic",
		Params:   make(map[string]ParamResult),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range anthropicParams {
		wg.Add(1)
		go func(param string, value any) {
			defer wg.Done()
			errMsg := testAnthropicWithParam(model, apiKey, param, value)
			supported := true
			if errMsg != "" && (strings.Contains(errMsg, "not supported") || strings.Contains(errMsg, "Unsupported") || strings.Contains(errMsg, "not_found_error")) {
				supported = false
			}

			mu.Lock()
			if strings.Contains(errMsg, "not_found_error") {
				result.Params[param] = ParamResult{Supported: false, Error: "model not found"}
			} else {
				result.Params[param] = ParamResult{Supported: supported, Error: errMsg}
			}
			mu.Unlock()
		}(p.name, p.value)
	}

	wg.Wait()
	return result
}

func testAnthropicWithParam(model, apiKey, param string, value any) string {
	payload := map[string]any{
		"model":      model,
		"max_tokens": 10,
		"messages": []map[string]string{
			{"role": "user", "content": "Say hi"},
		},
	}
	// Don't override max_tokens if that's the param we're testing
	if param != "max_tokens" {
		payload[param] = value
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return ""
	}

	respBody, _ := io.ReadAll(resp.Body)
	return string(respBody)
}

