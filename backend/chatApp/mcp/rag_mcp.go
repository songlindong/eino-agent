package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	componenttool "github.com/cloudwego/eino/components/tool"
	mcpclient "github.com/mark3labs/mcp-go/client"
	clienttransport "github.com/mark3labs/mcp-go/client/transport"
	mcpschema "github.com/mark3labs/mcp-go/mcp"
)

const (
	defaultMCPCommand     = "go"
	defaultClientName     = "eino-agent"
	defaultClientVersion  = "0.1.0"
	defaultRetrieveTool   = "rag.retrieve"
	ragRepoName           = "rag-retrievalOps"
	ragBackendDirName     = "backend"
	ragMCPServerEntryPath = "cmd/mcp-server/main.go"
)

type Config struct {
	BaseURL       string
	APIKey        string
	DefaultKBIDs  string
	TimeoutMS     string
	ServerName    string
	ServerVersion string

	WorkDir string
	Command string
	Args    []string

	ToolNames []string
}

type Toolset struct {
	Client mcpclient.MCPClient
	Tools  []componenttool.BaseTool
}

func LoadConfigFromEnv() (*Config, error) {
	workDir, err := resolveRAGBackendDir(strings.TrimSpace(os.Getenv("RAG_MCP_WORKDIR")))
	if err != nil {
		return nil, err
	}

	command := strings.TrimSpace(os.Getenv("RAG_MCP_COMMAND"))
	args := parseArgsEnv(os.Getenv("RAG_MCP_ARGS"))
	if command == "" {
		command = defaultMCPCommand
		args = []string{"run", "./cmd/mcp-server"}
	}

	cfg := &Config{
		BaseURL:       strings.TrimRight(strings.TrimSpace(os.Getenv("RAG_BASE_URL")), "/"),
		APIKey:        strings.TrimSpace(os.Getenv("RAG_API_KEY")),
		DefaultKBIDs:  strings.TrimSpace(os.Getenv("RAG_DEFAULT_KB_IDS")),
		TimeoutMS:     strings.TrimSpace(os.Getenv("RAG_TIMEOUT_MS")),
		ServerName:    strings.TrimSpace(os.Getenv("MCP_SERVER_NAME")),
		ServerVersion: strings.TrimSpace(os.Getenv("MCP_SERVER_VERSION")),
		WorkDir:       workDir,
		Command:       command,
		Args:          args,
		ToolNames:     []string{defaultRetrieveTool},
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("RAG_BASE_URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("RAG_API_KEY is required")
	}

	return cfg, nil
}

func NewToolset(ctx context.Context, cfg *Config) (*Toolset, error) {
	if cfg == nil {
		return nil, fmt.Errorf("rag mcp config is required")
	}
	if cfg.WorkDir == "" {
		return nil, fmt.Errorf("rag mcp workdir is required")
	}
	if cfg.Command == "" {
		return nil, fmt.Errorf("rag mcp command is required")
	}

	clientEnv := buildClientEnv(cfg)
	cli, err := mcpclient.NewStdioMCPClientWithOptions(
		cfg.Command,
		clientEnv,
		cfg.Args,
		clienttransport.WithCommandFunc(func(ctx context.Context, command string, env []string, args []string) (*exec.Cmd, error) {
			cmd := exec.CommandContext(ctx, command, args...)
			cmd.Dir = cfg.WorkDir
			cmd.Env = append(os.Environ(), env...)
			return cmd, nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create rag mcp client: %w", err)
	}

	initReq := mcpschema.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpschema.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpschema.Implementation{
		Name:    defaultClientName,
		Version: defaultClientVersion,
	}

	if _, err := cli.Initialize(ctx, initReq); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("initialize rag mcp client: %w", err)
	}

	toolNames := cfg.ToolNames
	if len(toolNames) == 0 {
		toolNames = []string{defaultRetrieveTool}
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{
		Cli:          cli,
		ToolNameList: toolNames,
	})
	if err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("load rag mcp tools: %w", err)
	}

	return &Toolset{
		Client: cli,
		Tools:  tools,
	}, nil
}

func (t *Toolset) Close() error {
	if t == nil || t.Client == nil {
		return nil
	}
	return t.Client.Close()
}

func buildClientEnv(cfg *Config) []string {
	env := []string{
		"RAG_BASE_URL=" + cfg.BaseURL,
		"RAG_API_KEY=" + cfg.APIKey,
	}

	appendIfPresent := func(key, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		env = append(env, key+"="+strings.TrimSpace(value))
	}

	appendIfPresent("RAG_DEFAULT_KB_IDS", cfg.DefaultKBIDs)
	appendIfPresent("RAG_TIMEOUT_MS", cfg.TimeoutMS)
	appendIfPresent("MCP_SERVER_NAME", cfg.ServerName)
	appendIfPresent("MCP_SERVER_VERSION", cfg.ServerVersion)

	return env
}

func parseArgsEnv(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}

func resolveRAGBackendDir(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Clean(explicit), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve rag mcp workdir: %w", err)
	}

	for current := filepath.Clean(wd); ; current = filepath.Dir(current) {
		candidates := []string{
			filepath.Join(current, ragRepoName, ragBackendDirName),
			filepath.Join(current, "..", ragRepoName, ragBackendDirName),
		}
		for _, candidate := range candidates {
			if looksLikeRAGBackend(candidate) {
				return filepath.Clean(candidate), nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}

	return "", fmt.Errorf("unable to locate %s/%s, set RAG_MCP_WORKDIR explicitly", ragRepoName, ragBackendDirName)
}

func looksLikeRAGBackend(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(ragMCPServerEntryPath)))
	if err != nil {
		return false
	}
	return !info.IsDir()
}
