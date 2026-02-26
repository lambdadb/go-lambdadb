package lambdadb

import (
	"context"
	"net/http"
	"time"

	"github.com/lambdadb/go-lambdadb/internal/config"
	"github.com/lambdadb/go-lambdadb/internal/hooks"
	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/retry"
)

// HTTPClient provides an interface for supplying the SDK with a custom HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// String returns a pointer to s.
func String(s string) *string { return &s }

// Bool returns a pointer to b.
func Bool(b bool) *bool { return &b }

// Int returns a pointer to i.
func Int(i int) *int { return &i }

// Int64 returns a pointer to i.
func Int64(i int64) *int64 { return &i }

// Float32 returns a pointer to f.
func Float32(f float32) *float32 { return &f }

// Float64 returns a pointer to f.
func Float64(f float64) *float64 { return &f }

// Pointer returns a pointer to v.
func Pointer[T any](v T) *T { return &v }

// Version is the SDK version. Use it for SDKVersion and User-Agent.
const Version = "0.2.1"

// Client is the LambdaDB API client.
type Client struct {
	SDKVersion string

	// Collections provides project-level collection operations (list and create).
	Collections *ProjectCollections

	// collections and docs are used internally by Collection and CollectionDocs.
	collections *Collections
	docs        *Docs

	sdkConfiguration config.SDKConfiguration
	hooks            *hooks.Hooks
}

// SDKOption configures the client.
type SDKOption func(*Client)

// WithBaseURL sets the API base URL (scheme + host, e.g. https://api.lambdadb.ai).
// Default is https://api.lambdadb.ai per OpenAPI spec.
func WithBaseURL(baseURL string) SDKOption {
	return func(c *Client) {
		c.sdkConfiguration.BaseURL = baseURL
	}
}

// WithProjectName sets the project name (path segment, e.g. playground).
// Default is "playground" per OpenAPI spec.
func WithProjectName(projectName string) SDKOption {
	return func(c *Client) {
		c.sdkConfiguration.ProjectName = projectName
	}
}

// WithAPIKey sets the project API key for authentication.
func WithAPIKey(apiKey string) SDKOption {
	return func(c *Client) {
		security := components.Security{ProjectAPIKey: &apiKey}
		c.sdkConfiguration.Security = utils.AsSecuritySource(&security)
	}
}

// WithClient sets the HTTP client used by the SDK.
func WithClient(client HTTPClient) SDKOption {
	return func(c *Client) {
		c.sdkConfiguration.Client = client
	}
}

// WithSecuritySource sets a function that provides security on each request.
func WithSecuritySource(security func(context.Context) (components.Security, error)) SDKOption {
	return func(c *Client) {
		c.sdkConfiguration.Security = func(ctx context.Context) (interface{}, error) {
			return security(ctx)
		}
	}
}

// WithRetryConfig sets the default retry configuration.
func WithRetryConfig(retryConfig retry.Config) SDKOption {
	return func(c *Client) {
		c.sdkConfiguration.RetryConfig = &retryConfig
	}
}

// WithTimeout sets the default request timeout.
func WithTimeout(timeout time.Duration) SDKOption {
	return func(c *Client) {
		c.sdkConfiguration.Timeout = &timeout
	}
}

// New creates a new LambdaDB client with the given options.
func New(opts ...SDKOption) *Client {
	c := &Client{
		SDKVersion: Version,
		sdkConfiguration: config.SDKConfiguration{
			UserAgent:   "lambdadb-go/" + Version,
			BaseURL:     config.DefaultBaseURL,
			ProjectName: config.DefaultProjectName,
		},
		hooks: hooks.New(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(c)
	}

	if c.sdkConfiguration.Security == nil {
		var envVarSecurity components.Security
		if utils.PopulateSecurityFromEnv(&envVarSecurity) {
			c.sdkConfiguration.Security = utils.AsSecuritySource(envVarSecurity)
		}
	}

	if c.sdkConfiguration.Client == nil {
		c.sdkConfiguration.Client = &http.Client{Timeout: 60 * time.Second}
	}

	c.sdkConfiguration = c.hooks.SDKInit(c.sdkConfiguration)

	c.collections = newCollections(c, c.sdkConfiguration, c.hooks)
	c.docs = newDocs(c, c.sdkConfiguration, c.hooks)
	c.Collections = &ProjectCollections{client: c}

	return c
}

// Collection returns a handle for the named collection.
// Use it for collection-level operations and document operations without passing the collection name each time.
func (c *Client) Collection(name string) *Collection {
	return &Collection{client: c, name: name}
}
