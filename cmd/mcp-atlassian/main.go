// Command mcp-atlassian runs an Atlassian MCP server that exposes tools for
// Jira and Confluence. It can also be used as a CLI tool for testing and scripting.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtime "github.com/plexusone/omniskill/mcp/server"
	"github.com/plexusone/omnitoken"
	"github.com/spf13/cobra"

	// Register desktop vault providers (1Password, etc.)
	_ "github.com/plexusone/omnivault-desktop"

	"github.com/grokify/go-atlassian/confluence"
	"github.com/grokify/go-atlassian/jira"
	confskill "github.com/plexusone/mcp-atlassian/skills/confluence"
	jiraskill "github.com/plexusone/mcp-atlassian/skills/jira"
)

const (
	serverName    = "mcp-atlassian"
	serverVersion = "v0.4.0"
)

var (
	// Credential flags (persistent across all commands).
	// baseURL is the Atlassian site URL (no /wiki suffix).
	baseURL         string
	username        string
	apiToken        string
	vaultURI        string
	credentialsName string

	// Output format flag
	outputFormat string

	// search-pages flags
	searchLimit int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mcp-atlassian",
	Short: "MCP server and CLI for Atlassian (Jira + Confluence)",
	Long: `An MCP (Model Context Protocol) server for Atlassian products including
Jira and Confluence. Can also be used as a CLI tool for testing and scripting.

Running without a subcommand starts the MCP server (default behavior).

Credentials can be provided via:
  - Direct credentials (base URL, username, API token)
  - Vault-backed credentials via omnitoken`,
	Example: `  # Start MCP server (default)
  mcp-atlassian --base-url https://example.atlassian.net \
                --username user@example.com --api-token your-token

  # Or use environment variables
  export ATLASSIAN_URL=https://example.atlassian.net
  export ATLASSIAN_USERNAME=user@example.com
  export ATLASSIAN_API_TOKEN=your-token
  mcp-atlassian

  # CLI: Read a Confluence page
  mcp-atlassian read-page 12345

  # CLI: Search for Confluence pages
  mcp-atlassian search-pages "type=page AND text~authentication"`,
	SilenceUsage: true,
	RunE:         runServer,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server (stdio)",
	Long:  "Start the MCP server using stdio transport for communication with MCP clients.",
	RunE:  runServer,
}

var (
	httpAddr       string
	httpPath       string
	oauthUser      string
	oauthPassword  string
	oauthClientID  string
	oauthClientSec string
	oauthDebug     bool
	ngrokAuthtoken string
	ngrokDomain    string
)

var serveHTTPCmd = &cobra.Command{
	Use:   "serve-http",
	Short: "Start the MCP server over HTTP with OAuth 2.1",
	Long: `Start the MCP server over HTTP with OAuth 2.1 Authorization Code + PKCE.

This mode is suitable for remote MCP clients (ChatGPT.com, web apps, shared
team deployments). Clients authenticate via OAuth 2.1 with PKCE before
accessing MCP tools.

Optionally expose the server via ngrok for public access.`,
	Example: `  # Local HTTP with OAuth
  mcp-atlassian serve-http --addr :8080 \
    --oauth-user admin --oauth-password secret

  # With ngrok tunnel
  mcp-atlassian serve-http --addr :8080 \
    --oauth-user admin --oauth-password secret \
    --ngrok-authtoken $NGROK_AUTHTOKEN

  # Pre-registered client for ChatGPT.com
  mcp-atlassian serve-http --addr :8080 \
    --oauth-user admin --oauth-password secret \
    --oauth-client-id my-client --oauth-client-secret my-secret`,
	RunE: runHTTPServer,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s\n", serverName, serverVersion)
	},
}

// CLI commands for Confluence tools
var readPageCmd = &cobra.Command{
	Use:   "read-page <page-id>",
	Short: "Read a page as structured blocks",
	Long:  "Read a Confluence page and return its content as structured blocks.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTool("read_page", map[string]any{
			"page_id": args[0],
		})
	},
}

var readPageXHTMLCmd = &cobra.Command{
	Use:   "read-page-xhtml <page-id>",
	Short: "Read a page as raw XHTML",
	Long:  "Read a Confluence page and return its content as raw XHTML.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTool("read_page_xhtml", map[string]any{
			"page_id": args[0],
		})
	},
}

var searchPagesCmd = &cobra.Command{
	Use:   "search-pages <cql>",
	Short: "Search pages using CQL",
	Long:  "Search for Confluence pages using CQL (Confluence Query Language).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := map[string]any{
			"cql": args[0],
		}
		if searchLimit > 0 {
			params["limit"] = searchLimit
		}
		return runTool("search_pages", params)
	},
}

var deletePageCmd = &cobra.Command{
	Use:   "delete-page <page-id>",
	Short: "Delete a page",
	Long:  "Delete a Confluence page by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTool("delete_page", map[string]any{
			"page_id": args[0],
		})
	},
}

func init() {
	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "",
		"Atlassian site URL, e.g. https://example.atlassian.net (env: ATLASSIAN_URL)")
	rootCmd.PersistentFlags().StringVar(&username, "username", "",
		"Atlassian username/email (env: ATLASSIAN_USERNAME or CONFLUENCE_USERNAME)")
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", "",
		"Atlassian API token (env: ATLASSIAN_API_TOKEN or CONFLUENCE_API_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&vaultURI, "vault", "",
		"vault URI for credentials (env: OMNITOKEN_VAULT_URI)")
	rootCmd.PersistentFlags().StringVar(&credentialsName, "credentials-name", "",
		"name of credentials in vault (env: OMNITOKEN_CREDENTIALS_NAME)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json",
		"output format: json, pretty (default: json)")

	// search-pages flags
	searchPagesCmd.Flags().IntVar(&searchLimit, "limit", 0, "maximum results to return")

	// serve-http flags
	serveHTTPCmd.Flags().StringVar(&httpAddr, "addr", ":8080", "HTTP listen address")
	serveHTTPCmd.Flags().StringVar(&httpPath, "path", "/mcp", "HTTP path for MCP endpoint")
	serveHTTPCmd.Flags().StringVar(&oauthUser, "oauth-user", "", "OAuth login username (env: MCP_OAUTH_USER)")
	serveHTTPCmd.Flags().StringVar(&oauthPassword, "oauth-password", "", "OAuth login password (env: MCP_OAUTH_PASSWORD)")
	serveHTTPCmd.Flags().StringVar(&oauthClientID, "oauth-client-id", "", "pre-registered OAuth client ID (auto-generated if empty)")
	serveHTTPCmd.Flags().StringVar(&oauthClientSec, "oauth-client-secret", "", "pre-registered OAuth client secret (auto-generated if empty)")
	serveHTTPCmd.Flags().BoolVar(&oauthDebug, "oauth-debug", false, "enable OAuth debug logging")
	serveHTTPCmd.Flags().StringVar(&ngrokAuthtoken, "ngrok-authtoken", "", "ngrok auth token for public tunneling (env: NGROK_AUTHTOKEN)")
	serveHTTPCmd.Flags().StringVar(&ngrokDomain, "ngrok-domain", "", "custom ngrok domain (requires paid plan)")

	// Add commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(serveHTTPCmd)
	rootCmd.AddCommand(versionCmd)

	// Confluence CLI commands
	rootCmd.AddCommand(readPageCmd)
	rootCmd.AddCommand(readPageXHTMLCmd)
	rootCmd.AddCommand(searchPagesCmd)
	rootCmd.AddCommand(deletePageCmd)
}

// applyEnvDefaults applies environment variable defaults to flags.
// ATLASSIAN_URL is the site root (e.g. https://example.atlassian.net).
// Legacy CONFLUENCE_BASE_URL (which includes /wiki) is normalized by stripping the suffix.
func applyEnvDefaults() {
	if baseURL == "" {
		if v := os.Getenv("ATLASSIAN_URL"); v != "" {
			baseURL = strings.TrimRight(v, "/")
		} else if v := os.Getenv("CONFLUENCE_BASE_URL"); v != "" {
			baseURL = strings.TrimRight(strings.TrimSuffix(v, "/wiki"), "/")
		}
	}
	if username == "" {
		username = firstNonEmpty(os.Getenv("ATLASSIAN_USERNAME"), os.Getenv("CONFLUENCE_USERNAME"))
	}
	if apiToken == "" {
		apiToken = firstNonEmpty(os.Getenv("ATLASSIAN_API_TOKEN"), os.Getenv("CONFLUENCE_API_TOKEN"))
	}
	if vaultURI == "" {
		vaultURI = os.Getenv("OMNITOKEN_VAULT_URI")
	}
	if credentialsName == "" {
		credentialsName = os.Getenv("OMNITOKEN_CREDENTIALS_NAME")
	}
	if credentialsName == "" {
		credentialsName = "atlassian"
	}
}

// confluenceURL returns the Confluence API base URL (baseURL + /wiki).
func confluenceURL() string {
	return baseURL + "/wiki"
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// skills holds initialized skills for the server.
type skills struct {
	confluence *confskill.Skill
	jira       *jiraskill.Skill
	cleanup    func()
}

// getSkills creates and initializes both Confluence and Jira skills with shared credentials.
func getSkills(ctx context.Context) (*skills, error) {
	applyEnvDefaults()

	hasDirectCreds := baseURL != "" && username != "" && apiToken != ""
	hasVaultCreds := vaultURI != ""

	var confClient *confluence.Client
	var jiraClient *jira.Client
	var tokenMgr *omnitoken.TokenManager

	cleanup := func() {}

	if hasVaultCreds {
		var err error
		tokenMgr, err = omnitoken.NewFromVaultURI(vaultURI)
		if err != nil {
			return nil, fmt.Errorf("failed to create token manager: %w", err)
		}
		cleanup = func() {
			if err := tokenMgr.Close(); err != nil {
				log.Printf("Warning: failed to close token manager: %v", err)
			}
		}

		httpClient, err := tokenMgr.GetClient(ctx, credentialsName)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to get authenticated client: %w", err)
		}

		if baseURL == "" {
			cleanup()
			return nil, fmt.Errorf("--base-url is required when using vault credentials")
		}

		confClient = confluence.NewClient(confluenceURL(), &bearerAuthFromClient{httpClient}, confluence.WithHTTPClient(httpClient))
		jiraClient, err = jira.NewClientFromHTTPClient(baseURL, httpClient)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to create Jira client: %w", err)
		}
	} else if hasDirectCreds {
		confClient = confluence.NewClient(confluenceURL(), confluence.BasicAuth{
			Username: username,
			Token:    apiToken,
		})
		var err error
		jiraClient, err = jira.NewClientFromBasicAuth(baseURL, username, apiToken, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create Jira client: %w", err)
		}
	} else {
		return nil, fmt.Errorf("credentials required: use --base-url/--username/--api-token or --vault")
	}

	confSkill := confskill.New(confClient)
	if err := confSkill.Init(ctx); err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to initialize Confluence skill: %w", err)
	}

	jiraSkill := jiraskill.New(jiraClient)
	if err := jiraSkill.Init(ctx); err != nil {
		if cerr := confSkill.Close(); cerr != nil {
			log.Printf("Warning: failed to close Confluence skill: %v", cerr)
		}
		cleanup()
		return nil, fmt.Errorf("failed to initialize Jira skill: %w", err)
	}

	s := &skills{
		confluence: confSkill,
		jira:       jiraSkill,
		cleanup: func() {
			if err := jiraSkill.Close(); err != nil {
				log.Printf("Warning: failed to close Jira skill: %v", err)
			}
			if err := confSkill.Close(); err != nil {
				log.Printf("Warning: failed to close Confluence skill: %v", err)
			}
			cleanup()
		},
	}

	return s, nil
}

// outputResult outputs the result in the specified format
func outputResult(result any) error {
	var data []byte
	var err error

	switch outputFormat {
	case "pretty":
		data, err = json.MarshalIndent(result, "", "  ")
	default:
		data, err = json.Marshal(result)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// runTool runs a tool by name with the given params (searches all skills).
func runTool(toolName string, params map[string]any) error {
	ctx := context.Background()

	s, err := getSkills(ctx)
	if err != nil {
		return err
	}
	defer s.cleanup()

	allTools := append(s.confluence.Tools(), s.jira.Tools()...)
	for _, tool := range allTools {
		if tool.Name() == toolName {
			result, err := tool.Call(ctx, params)
			if err != nil {
				return err
			}
			return outputResult(result)
		}
	}
	return fmt.Errorf("tool not found: %s", toolName)
}

func newRuntime(s *skills) *runtime.Runtime {
	rt := runtime.New(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)
	rt.RegisterSkill(s.confluence)
	rt.RegisterSkill(s.jira)
	return rt
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	s, err := getSkills(ctx)
	if err != nil {
		return err
	}
	defer s.cleanup()

	if err := newRuntime(s).ServeStdio(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func runHTTPServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Resolve OAuth credentials from flags or env
	user := firstNonEmpty(oauthUser, os.Getenv("MCP_OAUTH_USER"))
	pass := firstNonEmpty(oauthPassword, os.Getenv("MCP_OAUTH_PASSWORD"))
	if user == "" || pass == "" {
		return fmt.Errorf("--oauth-user and --oauth-password are required (or MCP_OAUTH_USER / MCP_OAUTH_PASSWORD)")
	}

	s, err := getSkills(ctx)
	if err != nil {
		return err
	}
	defer s.cleanup()

	rt := newRuntime(s)

	httpOpts := &runtime.HTTPServerOptions{
		Addr: httpAddr,
		Path: httpPath,
		OAuth2: &runtime.OAuth2Options{
			Users:        map[string]string{user: pass},
			ClientID:     oauthClientID,
			ClientSecret: oauthClientSec,
			Debug:        oauthDebug,
		},
		OnReady: func(result *runtime.HTTPServerResult) {
			log.Printf("MCP server listening at %s", result.LocalURL)
			if result.PublicURL != "" {
				log.Printf("Public URL: %s", result.PublicURL)
			}
			if result.OAuth2 != nil {
				log.Printf("OAuth2 client_id: %s", result.OAuth2.ClientID)
				log.Printf("OAuth2 authorization: %s", result.OAuth2.AuthorizationEndpoint)
				log.Printf("OAuth2 token: %s", result.OAuth2.TokenEndpoint)
				log.Printf("OAuth2 registration: %s", result.OAuth2.RegistrationEndpoint)
			}
		},
	}

	// Configure ngrok if requested
	authtoken := firstNonEmpty(ngrokAuthtoken, os.Getenv("NGROK_AUTHTOKEN"))
	if authtoken != "" {
		httpOpts.Ngrok = &runtime.NgrokOptions{
			Authtoken: authtoken,
			Domain:    ngrokDomain,
		}
	}

	if _, err := rt.ServeHTTP(ctx, httpOpts); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}

// bearerAuthFromClient is a no-op auth method when the HTTP client already has auth.
type bearerAuthFromClient struct {
	client *http.Client
}

func (b *bearerAuthFromClient) Apply(req *http.Request) {
	// No-op: the HTTP client already handles authentication
}
