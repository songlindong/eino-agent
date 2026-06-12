package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	componenttool "github.com/cloudwego/eino/components/tool"
)

type fakeRetrieveRequest struct {
	Query           string   `json:"query"`
	KBIDs           []uint64 `json:"kb_ids"`
	TopK            int      `json:"top_k"`
	StrategyProfile string   `json:"strategy_profile"`
}

func TestNewToolsetCallsRealRAGMCPServer(t *testing.T) {
	t.Parallel()

	var (
		gotAuthHeader string
		gotRequest    fakeRetrieveRequest
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/retrieve" {
			http.NotFound(w, r)
			return
		}

		gotAuthHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode retrieve request failed: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    200,
			"message": "Success",
			"data": map[string]any{
				"request_id": "req-test-1",
				"items": []map[string]any{
					{
						"content": "RAG 是检索增强生成，用于先检索外部知识再让模型作答。",
						"score":   0.99,
						"citation": map[string]any{
							"kb_id": 42,
						},
						"source": map[string]any{
							"route": "dense",
						},
					},
				},
			},
		})
	}))
	defer upstream.Close()

	workDir, err := resolveRAGBackendDir("")
	if err != nil {
		t.Fatalf("resolve rag backend dir failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	toolset, err := NewToolset(ctx, &Config{
		BaseURL:      upstream.URL,
		APIKey:       "rag_test_key",
		DefaultKBIDs: "42",
		WorkDir:      workDir,
		Command:      "go",
		Args:         []string{"run", "./cmd/mcp-server"},
		ToolNames:    []string{defaultRetrieveTool},
	})
	if err != nil {
		t.Fatalf("new toolset failed: %v", err)
	}
	defer func() {
		if err := toolset.Close(); err != nil {
			t.Fatalf("close toolset failed: %v", err)
		}
	}()

	if len(toolset.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(toolset.Tools))
	}

	info, err := toolset.Tools[0].Info(ctx)
	if err != nil {
		t.Fatalf("read tool info failed: %v", err)
	}
	if info.Name != defaultRetrieveTool {
		t.Fatalf("unexpected tool name: %s", info.Name)
	}

	invokable, ok := toolset.Tools[0].(componenttool.InvokableTool)
	if !ok {
		t.Fatalf("rag mcp tool does not implement InvokableTool")
	}

	result, err := invokable.InvokableRun(ctx, `{"query":"什么是RAG","top_k":2}`)
	if err != nil {
		t.Fatalf("invoke rag.retrieve failed: %v", err)
	}

	if gotAuthHeader != "Bearer rag_test_key" {
		t.Fatalf("unexpected auth header: %s", gotAuthHeader)
	}
	if gotRequest.Query != "什么是RAG" {
		t.Fatalf("unexpected query: %s", gotRequest.Query)
	}
	if len(gotRequest.KBIDs) != 1 || gotRequest.KBIDs[0] != 42 {
		t.Fatalf("unexpected kb_ids: %#v", gotRequest.KBIDs)
	}
	if gotRequest.TopK != 2 {
		t.Fatalf("unexpected top_k: %d", gotRequest.TopK)
	}
	if !strings.Contains(result, `"request_id":"req-test-1"`) {
		t.Fatalf("result missing request id: %s", result)
	}
	if !strings.Contains(result, `"content":"RAG 是检索增强生成`) {
		t.Fatalf("result missing retrieved content: %s", result)
	}
}
