package lambdadb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lambdadb/go-lambdadb/internal/hooks"
	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"github.com/lambdadb/go-lambdadb/retry"
)

// CreateBranchInput configures a new writable branch. Source may be omitted to
// create the branch from main.
type CreateBranchInput struct {
	BranchName string     `json:"branchName"`
	Source     *RefSource `json:"source,omitempty"`
}

// CreateTagInput configures a new immutable tag. Source may be omitted to
// create the tag from main.
type CreateTagInput struct {
	TagName string     `json:"tagName"`
	Source  *RefSource `json:"source,omitempty"`
}

// CreateAliasInput configures a new alias.
type CreateAliasInput struct {
	AliasName string      `json:"aliasName"`
	Target    AliasTarget `json:"target"`
}

// RetargetAliasInput configures the new target of an existing alias.
type RetargetAliasInput struct {
	Target AliasTarget `json:"target"`
}

// CollectionBranches provides branch operations for one collection.
type CollectionBranches struct {
	collection *Collection
}

// Create creates a writable branch.
func (b *CollectionBranches) Create(ctx context.Context, input CreateBranchInput, opts ...operations.Option) (*RefDetails, error) {
	var response struct {
		Branch RefDetails `json:"branch"`
	}
	err := b.collection.client.doVersioningRequest(
		ctx,
		http.MethodPost,
		"createBranch",
		[]string{"collections", b.collection.name, "branches"},
		input,
		http.StatusCreated,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Branch, nil
}

// List lists all branches in the collection.
func (b *CollectionBranches) List(ctx context.Context, opts ...operations.Option) ([]RefDetails, error) {
	var response struct {
		Branches []RefDetails `json:"branches"`
	}
	err := b.collection.client.doVersioningRequest(
		ctx,
		http.MethodGet,
		"listBranches",
		[]string{"collections", b.collection.name, "branches"},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return response.Branches, nil
}

// Delete deletes a branch. The default main branch cannot be deleted.
func (b *CollectionBranches) Delete(ctx context.Context, branchName string, opts ...operations.Option) (*MessageResponse, error) {
	var response MessageResponse
	err := b.collection.client.doVersioningRequest(
		ctx,
		http.MethodDelete,
		"deleteBranch",
		[]string{"collections", b.collection.name, "branches", branchName},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CollectionTags provides tag operations for one collection.
type CollectionTags struct {
	collection *Collection
}

// Create creates an immutable tag.
func (t *CollectionTags) Create(ctx context.Context, input CreateTagInput, opts ...operations.Option) (*RefDetails, error) {
	var response struct {
		Tag RefDetails `json:"tag"`
	}
	err := t.collection.client.doVersioningRequest(
		ctx,
		http.MethodPost,
		"createTag",
		[]string{"collections", t.collection.name, "tags"},
		input,
		http.StatusCreated,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Tag, nil
}

// List lists all tags in the collection.
func (t *CollectionTags) List(ctx context.Context, opts ...operations.Option) ([]RefDetails, error) {
	var response struct {
		Tags []RefDetails `json:"tags"`
	}
	err := t.collection.client.doVersioningRequest(
		ctx,
		http.MethodGet,
		"listTags",
		[]string{"collections", t.collection.name, "tags"},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return response.Tags, nil
}

// Delete deletes a tag.
func (t *CollectionTags) Delete(ctx context.Context, tagName string, opts ...operations.Option) (*MessageResponse, error) {
	var response MessageResponse
	err := t.collection.client.doVersioningRequest(
		ctx,
		http.MethodDelete,
		"deleteTag",
		[]string{"collections", t.collection.name, "tags", tagName},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CollectionAliases provides alias operations for one collection.
type CollectionAliases struct {
	collection *Collection
}

// Create creates an alias targeting a branch or tag.
func (a *CollectionAliases) Create(ctx context.Context, input CreateAliasInput, opts ...operations.Option) (*AliasDetails, error) {
	var response struct {
		Alias AliasDetails `json:"alias"`
	}
	err := a.collection.client.doVersioningRequest(
		ctx,
		http.MethodPost,
		"createAlias",
		[]string{"collections", a.collection.name, "aliases"},
		input,
		http.StatusCreated,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Alias, nil
}

// List lists all aliases in the collection.
func (a *CollectionAliases) List(ctx context.Context, opts ...operations.Option) ([]AliasDetails, error) {
	var response struct {
		Aliases []AliasDetails `json:"aliases"`
	}
	err := a.collection.client.doVersioningRequest(
		ctx,
		http.MethodGet,
		"listAliases",
		[]string{"collections", a.collection.name, "aliases"},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return response.Aliases, nil
}

// Retarget changes an alias target.
func (a *CollectionAliases) Retarget(ctx context.Context, aliasName string, input RetargetAliasInput, opts ...operations.Option) (*AliasDetails, error) {
	var response struct {
		Alias AliasDetails `json:"alias"`
	}
	err := a.collection.client.doVersioningRequest(
		ctx,
		http.MethodPatch,
		"retargetAlias",
		[]string{"collections", a.collection.name, "aliases", aliasName},
		input,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Alias, nil
}

// Delete deletes an alias.
func (a *CollectionAliases) Delete(ctx context.Context, aliasName string, opts ...operations.Option) (*MessageResponse, error) {
	var response MessageResponse
	err := a.collection.client.doVersioningRequest(
		ctx,
		http.MethodDelete,
		"deleteAlias",
		[]string{"collections", a.collection.name, "aliases", aliasName},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) doVersioningRequest(
	ctx context.Context,
	method string,
	operationID string,
	pathSegments []string,
	body any,
	successStatus int,
	out any,
	opts ...operations.Option,
) error {
	o := operations.Options{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&o, operations.SupportedOptionRetries, operations.SupportedOptionTimeout); err != nil {
			return fmt.Errorf("error applying option: %w", err)
		}
	}

	baseURL := c.sdkConfiguration.GetServerDetailsURL()
	if o.ServerURL != nil {
		baseURL = *o.ServerURL
	}
	opURL, err := url.JoinPath(baseURL, pathSegments...)
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	var bodyReader *bytes.Reader
	if body == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error serializing request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	timeout := o.Timeout
	if timeout == nil {
		timeout = c.sdkConfiguration.Timeout
	}
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, opURL, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.sdkConfiguration.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := utils.PopulateSecurity(ctx, req, c.sdkConfiguration.Security); err != nil {
		return err
	}
	for key, value := range o.SetHeaders {
		req.Header.Set(key, value)
	}

	hookCtx := hooks.HookContext{
		SDK:              c,
		SDKConfiguration: c.sdkConfiguration,
		BaseURL:          baseURL,
		Context:          ctx,
		OperationID:      operationID,
		SecuritySource:   c.sdkConfiguration.Security,
	}

	retryConfig := o.Retries
	if retryConfig == nil {
		retryConfig = c.sdkConfiguration.RetryConfig
	}
	if retryConfig == nil {
		retryConfig = &retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 500,
				MaxInterval:     60000,
				Exponent:        1.5,
				MaxElapsedTime:  3600000,
			},
			RetryConnectionErrors: true,
		}
	}

	httpRes, err := utils.Retry(ctx, utils.Retries{
		Config:      retryConfig,
		StatusCodes: []string{"429", "5XX"},
	}, func() (*http.Response, error) {
		if req.GetBody != nil {
			req.Body, err = req.GetBody()
			if err != nil {
				return nil, err
			}
		}

		hookedReq, err := c.hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
		if err != nil {
			if retry.IsPermanentError(err) || retry.IsTemporaryError(err) {
				return nil, err
			}
			return nil, retry.Permanent(err)
		}
		req = hookedReq

		response, err := c.sdkConfiguration.Client.Do(req)
		if err != nil || response == nil {
			if err == nil {
				err = fmt.Errorf("error sending request: no response")
			} else {
				err = fmt.Errorf("error sending request: %w", err)
			}
			_, err = c.hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		}
		return response, err
	})
	if err != nil {
		return err
	}
	if httpRes == nil {
		return fmt.Errorf("error sending request: no response")
	}

	if httpRes.StatusCode != successStatus {
		hookedRes, hookErr := c.hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if hookErr != nil {
			return hookErr
		}
		if hookedRes != nil {
			httpRes = hookedRes
		}
		return decodeVersioningError(req, httpRes)
	}

	httpRes, err = c.hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
	if err != nil {
		return err
	}
	if !utils.MatchContentType(httpRes.Header.Get("Content-Type"), "application/json") {
		rawBody, readErr := utils.ConsumeRawBody(httpRes)
		if readErr != nil {
			return readErr
		}
		return apierrors.NewAPIError(
			fmt.Sprintf("unknown content-type received: %s", httpRes.Header.Get("Content-Type")),
			httpRes.StatusCode,
			string(rawBody),
			httpRes,
		)
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rawBody, out); err != nil {
		return fmt.Errorf("error decoding %s response: %w", operationID, err)
	}
	return nil
}

func decodeVersioningError(req *http.Request, response *http.Response) error {
	rawBody, err := utils.ConsumeRawBody(response)
	if err != nil {
		return err
	}
	metadata := components.HTTPMetadata{Request: req, Response: response}

	switch response.StatusCode {
	case http.StatusBadRequest:
		out := &apierrors.BadRequestError{HTTPMeta: metadata}
		if err := json.Unmarshal(rawBody, out); err == nil {
			return out
		}
	case http.StatusUnauthorized:
		out := &apierrors.UnauthenticatedError{HTTPMeta: metadata}
		if err := json.Unmarshal(rawBody, out); err == nil {
			return out
		}
	case http.StatusNotFound:
		out := &apierrors.ResourceNotFoundError{HTTPMeta: metadata}
		if err := json.Unmarshal(rawBody, out); err == nil {
			return out
		}
	case http.StatusConflict:
		out := &apierrors.ResourceAlreadyExistsError{HTTPMeta: metadata}
		if err := json.Unmarshal(rawBody, out); err == nil {
			return out
		}
	case http.StatusTooManyRequests:
		out := &apierrors.TooManyRequestsError{HTTPMeta: metadata}
		if err := json.Unmarshal(rawBody, out); err == nil {
			return out
		}
	default:
		if response.StatusCode >= 500 && response.StatusCode <= 599 {
			out := &apierrors.InternalServerError{HTTPMeta: metadata}
			if err := json.Unmarshal(rawBody, out); err == nil {
				return out
			}
		}
	}

	return apierrors.NewAPIError("API error occurred", response.StatusCode, string(rawBody), response)
}
