package config

import (
	"testing"
)

func TestGetServerDetails_Defaults(t *testing.T) {
	c := &SDKConfiguration{
		BaseURL:     DefaultBaseURL,
		ProjectName: DefaultProjectName,
	}
	url, vars := c.GetServerDetails()
	if url != "https://api.lambdadb.ai/projects/playground" {
		t.Errorf("GetServerDetails() url = %q, want https://api.lambdadb.ai/projects/playground", url)
	}
	if vars != nil {
		t.Errorf("GetServerDetails() vars = %v, want nil", vars)
	}
}

func TestGetServerDetails_Custom(t *testing.T) {
	c := &SDKConfiguration{
		BaseURL:     "https://api.example.com",
		ProjectName: "my-project",
	}
	url, _ := c.GetServerDetails()
	want := "https://api.example.com/projects/my-project"
	if url != want {
		t.Errorf("GetServerDetails() url = %q, want %q", url, want)
	}
}

func TestGetServerDetails_EmptyProjectUsesDefault(t *testing.T) {
	c := &SDKConfiguration{
		BaseURL:     "https://api.lambdadb.ai",
		ProjectName: "",
	}
	url, _ := c.GetServerDetails()
	if url != "https://api.lambdadb.ai/projects/playground" {
		t.Errorf("GetServerDetails() with empty project url = %q, want .../playground", url)
	}
}

func TestGetServerDetails_TrimTrailingSlash(t *testing.T) {
	c := &SDKConfiguration{
		BaseURL:     "https://api.lambdadb.ai/",
		ProjectName: "prod",
	}
	url, _ := c.GetServerDetails()
	if url != "https://api.lambdadb.ai/projects/prod" {
		t.Errorf("GetServerDetails() url = %q, want .../projects/prod", url)
	}
}

func TestGetServerDetailsURL(t *testing.T) {
	c := &SDKConfiguration{
		BaseURL:     "https://api.lambdadb.ai",
		ProjectName: "test",
	}
	url := c.GetServerDetailsURL()
	if url != "https://api.lambdadb.ai/projects/test" {
		t.Errorf("GetServerDetailsURL() = %q, want https://api.lambdadb.ai/projects/test", url)
	}
}
