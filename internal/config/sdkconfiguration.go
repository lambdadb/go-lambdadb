package config

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/lambdadb/go-lambdadb/retry"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SDKConfiguration holds the client and base URL configuration.
// BaseURL is the scheme + host (e.g. https://api.lambdadb.ai).
// ProjectName is the project path segment (e.g. playground).
// The effective request base is BaseURL + "/projects/" + ProjectName.
type SDKConfiguration struct {
	Client      HTTPClient
	Security    func(context.Context) (interface{}, error)
	BaseURL     string
	ProjectName string
	UserAgent   string
	RetryConfig *retry.Config
	Timeout     *time.Duration
}

// DefaultBaseURL is the default API base URL (OpenAPI spec default).
const DefaultBaseURL = "https://api.lambdadb.ai"

// DefaultProjectName is the default project name (OpenAPI spec default).
const DefaultProjectName = "playground"

// GetServerDetails returns the effective base URL for API requests,
// i.e. BaseURL + "/projects/" + ProjectName (no trailing slash).
func (c *SDKConfiguration) GetServerDetails() (string, map[string]string) {
	base := strings.TrimSuffix(c.BaseURL, "/")
	project := strings.Trim(c.ProjectName, "/")
	if project == "" {
		project = DefaultProjectName
	}
	return base + "/projects/" + project, nil
}

// GetServerDetailsURL returns the effective base URL as a string.
// Use this when you need a single string (e.g. for url.JoinPath).
func (c *SDKConfiguration) GetServerDetailsURL() string {
	s, _ := c.GetServerDetails()
	return s
}
