package lambdadb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/lambdadb/go-lambdadb/internal/hooks"
	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"github.com/lambdadb/go-lambdadb/retry"
)

// CreateBranchInput configures a new writable branch. Source may be omitted to
// create the branch from main. An explicit Source may be a branch or tag in the
// same collection; AsOf is valid only for a branch source. Alias and unknown
// source kinds are rejected locally.
type CreateBranchInput struct {
	BranchName string     `json:"branchName"`
	Source     *RefSource `json:"source,omitempty"`
}

// UpdateBranchInput updates the index schema of one branch. IndexConfigs is
// required and nonempty: it must preserve the branch's complete existing
// schema and may add fields at any object depth. The server-managed id field
// may be sent back as returned or omitted.
type UpdateBranchInput struct {
	IndexConfigs map[string]components.IndexConfigsUnion `json:"indexConfigs"`
}

// CreateTagInput configures a new immutable tag. Source may be omitted to
// create the tag from main. An explicit Source may be a branch or tag in the
// same collection; AsOf is valid only for a branch source.
type CreateTagInput struct {
	TagName string     `json:"tagName"`
	Source  *RefSource `json:"source,omitempty"`
}

// ListRefsOpts holds optional parameters for one page of a branch, tag, or
// alias listing. Pass nil to use defaults.
type ListRefsOpts struct {
	// Max number of entries to return at once, from 1 to 100. The server
	// default is 10. Nil sends no size.
	Size *int64
	// Opaque token returned by the previous page. Nil or empty sends no token,
	// which selects the first page; the server rejects an empty token.
	PageToken *string
}

// ListBranchesResult is one page of a branch listing.
type ListBranchesResult struct {
	Branches []BranchDetails
	// Opaque token for the next page; nil on the last page.
	NextPageToken *string
}

// ListTagsResult is one page of a tag listing.
type ListTagsResult struct {
	Tags []TagDetails
	// Opaque token for the next page; nil on the last page.
	NextPageToken *string
}

// ListAliasesResult is one page of an alias listing.
type ListAliasesResult struct {
	Aliases []AliasDetails
	// Opaque token for the next page; nil on the last page.
	NextPageToken *string
}

// refPager tracks the position of a ref list iterator.
type refPager struct {
	size      *int64
	nextToken *string
	done      bool
}

func newRefPager(listOpts *ListRefsOpts) refPager {
	pager := refPager{}
	if listOpts != nil {
		pager.size = listOpts.Size
		if listOpts.PageToken != nil {
			tok := *listOpts.PageToken
			pager.nextToken = &tok
		}
	}
	return pager
}

func (p *refPager) opts() *ListRefsOpts {
	if p.size == nil && p.nextToken == nil {
		return nil
	}
	return &ListRefsOpts{Size: p.size, PageToken: p.nextToken}
}

func (p *refPager) advance(nextPageToken *string) {
	if nextPageToken == nil || *nextPageToken == "" {
		p.done = true
	} else {
		p.nextToken = nextPageToken
	}
}

// refListQuery builds the query of a ref listing. Unset values and an empty
// page token are omitted: the server rejects an empty pageToken.
func refListQuery(listOpts *ListRefsOpts) url.Values {
	query := url.Values{}
	if listOpts == nil {
		return query
	}
	if listOpts.Size != nil {
		query.Set("size", strconv.FormatInt(*listOpts.Size, 10))
	}
	if listOpts.PageToken != nil && *listOpts.PageToken != "" {
		query.Set("pageToken", *listOpts.PageToken)
	}
	return query
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

// Create creates a writable branch from a branch or tag. Without AsOf the new
// branch starts from the source's committed snapshot and takes over the writes
// still pending on the source at creation; ordinary reads see them after the
// branch's own build commits them. A branch created from a tag inherits the
// tag's pinned schema and captured writes and has a nil ParentBranch.
func (b *CollectionBranches) Create(ctx context.Context, input CreateBranchInput, opts ...operations.Option) (*BranchDetails, error) {
	if input.Source != nil && input.Source.Kind != RefSourceKindBranch && input.Source.Kind != RefSourceKindTag {
		return nil, fmt.Errorf("branch source must be a branch or tag")
	}

	var response struct {
		Branch BranchDetails `json:"branch"`
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

// Get returns one branch with its index schema and data statistics. A name
// that is not a branch returns ResourceNotFoundError.
func (b *CollectionBranches) Get(ctx context.Context, branchName string, opts ...operations.Option) (*BranchDetails, error) {
	var response struct {
		Branch BranchDetails `json:"branch"`
	}
	err := b.collection.client.doVersioningRequest(
		ctx,
		http.MethodGet,
		"getBranch",
		[]string{"collections", b.collection.name, "branches", branchName},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Branch, nil
}

// Update updates the index schema of one branch and returns the updated
// branch. The change applies to this branch only; updating main also updates
// the collection's IndexConfigs, which mirror main. Naming a tag or an alias
// returns BadRequestError.
func (b *CollectionBranches) Update(ctx context.Context, branchName string, input UpdateBranchInput, opts ...operations.Option) (*BranchDetails, error) {
	var response struct {
		Branch BranchDetails `json:"branch"`
	}
	err := b.collection.client.doVersioningRequest(
		ctx,
		http.MethodPatch,
		"updateBranch",
		[]string{"collections", b.collection.name, "branches", branchName},
		input,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Branch, nil
}

// List lists all branches in the collection, in name order. The listing is
// paginated by the server; List fetches every page with the server's default
// page size. Use ListPage or ListIterator to control paging.
func (b *CollectionBranches) List(ctx context.Context, opts ...operations.Option) ([]BranchDetails, error) {
	it := b.ListIterator(ctx, nil, opts...)
	var out []BranchDetails
	for {
		page, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if page == nil {
			return out, nil
		}
		if out == nil {
			out = page.Branches
		} else {
			out = append(out, page.Branches...)
		}
	}
}

// ListPage returns one page of branches. Pass nil for listOpts to use
// defaults (no size or page token is sent).
func (b *CollectionBranches) ListPage(ctx context.Context, listOpts *ListRefsOpts, opts ...operations.Option) (*ListBranchesResult, error) {
	var response struct {
		Branches      []BranchDetails `json:"branches"`
		NextPageToken *string         `json:"nextPageToken"`
	}
	err := b.collection.client.doVersioningQueryRequest(
		ctx,
		http.MethodGet,
		"listBranches",
		[]string{"collections", b.collection.name, "branches"},
		refListQuery(listOpts),
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &ListBranchesResult{Branches: response.Branches, NextPageToken: response.NextPageToken}, nil
}

// BranchListIterator iterates over branch list pages without managing page
// tokens manually. Create it with ListIterator; then call Next until it
// returns (nil, nil).
type BranchListIterator struct {
	branches *CollectionBranches
	pager    refPager
	callOpts []operations.Option
}

// ListIterator returns an iterator that fetches branch list pages. Pass nil
// for listOpts to use defaults.
func (b *CollectionBranches) ListIterator(ctx context.Context, listOpts *ListRefsOpts, opts ...operations.Option) *BranchListIterator {
	return &BranchListIterator{branches: b, pager: newRefPager(listOpts), callOpts: opts}
}

// Next fetches the next page. It returns (result, nil), (nil, error), or
// (nil, nil) when there are no more pages.
func (it *BranchListIterator) Next(ctx context.Context) (*ListBranchesResult, error) {
	if it.pager.done {
		return nil, nil
	}
	page, err := it.branches.ListPage(ctx, it.pager.opts(), it.callOpts...)
	if err != nil {
		return nil, err
	}
	it.pager.advance(page.NextPageToken)
	return page, nil
}

// Delete deletes a branch. Deleting main returns BadRequestError (400).
// An alias-referenced branch returns ResourceAlreadyExistsError (409); delete
// or retarget every referencing alias before retrying.
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

// Create creates an immutable tag from a branch or tag. Without AsOf the tag
// pins the source's committed snapshot together with the writes still pending
// on the source at that moment; query and fetch reads of the tag include those
// captured writes, except pending bulk imports. With AsOf the tag pins only the
// selected committed snapshot.
func (t *CollectionTags) Create(ctx context.Context, input CreateTagInput, opts ...operations.Option) (*TagDetails, error) {
	var response struct {
		Tag TagDetails `json:"tag"`
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

// Get returns one tag with the index schema and data statistics of the
// snapshot it pins. A name that is not a tag returns ResourceNotFoundError.
func (t *CollectionTags) Get(ctx context.Context, tagName string, opts ...operations.Option) (*TagDetails, error) {
	var response struct {
		Tag TagDetails `json:"tag"`
	}
	err := t.collection.client.doVersioningRequest(
		ctx,
		http.MethodGet,
		"getTag",
		[]string{"collections", t.collection.name, "tags", tagName},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Tag, nil
}

// List lists all tags in the collection, in name order. The listing is
// paginated by the server; List fetches every page with the server's default
// page size. Use ListPage or ListIterator to control paging.
func (t *CollectionTags) List(ctx context.Context, opts ...operations.Option) ([]TagDetails, error) {
	it := t.ListIterator(ctx, nil, opts...)
	var out []TagDetails
	for {
		page, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if page == nil {
			return out, nil
		}
		if out == nil {
			out = page.Tags
		} else {
			out = append(out, page.Tags...)
		}
	}
}

// ListPage returns one page of tags. Pass nil for listOpts to use defaults
// (no size or page token is sent).
func (t *CollectionTags) ListPage(ctx context.Context, listOpts *ListRefsOpts, opts ...operations.Option) (*ListTagsResult, error) {
	var response struct {
		Tags          []TagDetails `json:"tags"`
		NextPageToken *string      `json:"nextPageToken"`
	}
	err := t.collection.client.doVersioningQueryRequest(
		ctx,
		http.MethodGet,
		"listTags",
		[]string{"collections", t.collection.name, "tags"},
		refListQuery(listOpts),
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &ListTagsResult{Tags: response.Tags, NextPageToken: response.NextPageToken}, nil
}

// TagListIterator iterates over tag list pages without managing page tokens
// manually. Create it with ListIterator; then call Next until it returns
// (nil, nil).
type TagListIterator struct {
	tags     *CollectionTags
	pager    refPager
	callOpts []operations.Option
}

// ListIterator returns an iterator that fetches tag list pages. Pass nil for
// listOpts to use defaults.
func (t *CollectionTags) ListIterator(ctx context.Context, listOpts *ListRefsOpts, opts ...operations.Option) *TagListIterator {
	return &TagListIterator{tags: t, pager: newRefPager(listOpts), callOpts: opts}
}

// Next fetches the next page. It returns (result, nil), (nil, error), or
// (nil, nil) when there are no more pages.
func (it *TagListIterator) Next(ctx context.Context) (*ListTagsResult, error) {
	if it.pager.done {
		return nil, nil
	}
	page, err := it.tags.ListPage(ctx, it.pager.opts(), it.callOpts...)
	if err != nil {
		return nil, err
	}
	it.pager.advance(page.NextPageToken)
	return page, nil
}

// Delete deletes a tag. An alias-referenced tag returns
// ResourceAlreadyExistsError (409); delete or retarget every referencing alias
// before retrying.
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

// Get returns one alias with the index schema it currently serves. A dangling
// alias is still returned, with Dangling true and nil IndexConfigs. A name
// that is not an alias returns ResourceNotFoundError.
func (a *CollectionAliases) Get(ctx context.Context, aliasName string, opts ...operations.Option) (*AliasDetails, error) {
	var response struct {
		Alias AliasDetails `json:"alias"`
	}
	err := a.collection.client.doVersioningRequest(
		ctx,
		http.MethodGet,
		"getAlias",
		[]string{"collections", a.collection.name, "aliases", aliasName},
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &response.Alias, nil
}

// List lists all aliases in the collection, in name order. The listing is
// paginated by the server; List fetches every page with the server's default
// page size. Use ListPage or ListIterator to control paging.
func (a *CollectionAliases) List(ctx context.Context, opts ...operations.Option) ([]AliasDetails, error) {
	it := a.ListIterator(ctx, nil, opts...)
	var out []AliasDetails
	for {
		page, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if page == nil {
			return out, nil
		}
		if out == nil {
			out = page.Aliases
		} else {
			out = append(out, page.Aliases...)
		}
	}
}

// ListPage returns one page of aliases. Pass nil for listOpts to use defaults
// (no size or page token is sent).
func (a *CollectionAliases) ListPage(ctx context.Context, listOpts *ListRefsOpts, opts ...operations.Option) (*ListAliasesResult, error) {
	var response struct {
		Aliases       []AliasDetails `json:"aliases"`
		NextPageToken *string        `json:"nextPageToken"`
	}
	err := a.collection.client.doVersioningQueryRequest(
		ctx,
		http.MethodGet,
		"listAliases",
		[]string{"collections", a.collection.name, "aliases"},
		refListQuery(listOpts),
		nil,
		http.StatusOK,
		&response,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &ListAliasesResult{Aliases: response.Aliases, NextPageToken: response.NextPageToken}, nil
}

// AliasListIterator iterates over alias list pages without managing page
// tokens manually. Create it with ListIterator; then call Next until it
// returns (nil, nil).
type AliasListIterator struct {
	aliases  *CollectionAliases
	pager    refPager
	callOpts []operations.Option
}

// ListIterator returns an iterator that fetches alias list pages. Pass nil for
// listOpts to use defaults.
func (a *CollectionAliases) ListIterator(ctx context.Context, listOpts *ListRefsOpts, opts ...operations.Option) *AliasListIterator {
	return &AliasListIterator{aliases: a, pager: newRefPager(listOpts), callOpts: opts}
}

// Next fetches the next page. It returns (result, nil), (nil, error), or
// (nil, nil) when there are no more pages.
func (it *AliasListIterator) Next(ctx context.Context) (*ListAliasesResult, error) {
	if it.pager.done {
		return nil, nil
	}
	page, err := it.aliases.ListPage(ctx, it.pager.opts(), it.callOpts...)
	if err != nil {
		return nil, err
	}
	it.pager.advance(page.NextPageToken)
	return page, nil
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
	return c.doVersioningQueryRequest(ctx, method, operationID, pathSegments, nil, body, successStatus, out, opts...)
}

func (c *Client) doVersioningQueryRequest(
	ctx context.Context,
	method string,
	operationID string,
	pathSegments []string,
	query url.Values,
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
	if len(query) > 0 {
		opURL += "?" + query.Encode()
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
