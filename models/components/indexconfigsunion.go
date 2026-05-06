package components

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lambdadb/go-lambdadb/internal/utils"
)

type TypeObject string

const (
	TypeObjectObject TypeObject = "object"
)

func (e TypeObject) ToPointer() *TypeObject {
	return &e
}
func (e *TypeObject) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "object":
		*e = TypeObject(v)
		return nil
	default:
		return fmt.Errorf("invalid value for TypeObject: %v", v)
	}
}

type IndexConfigsObject struct {
	Type               TypeObject     `json:"type"`
	ObjectIndexConfigs map[string]any `json:"objectIndexConfigs"`
}

func (i IndexConfigsObject) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(i, "", false)
}

func (i *IndexConfigsObject) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &i, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (i *IndexConfigsObject) GetType() TypeObject {
	if i == nil {
		return TypeObject("")
	}
	return i.Type
}

func (i *IndexConfigsObject) GetObjectIndexConfigs() map[string]any {
	if i == nil {
		return map[string]any{}
	}
	return i.ObjectIndexConfigs
}

type Type string

const (
	TypeKeyword      Type = "keyword"
	TypeLong         Type = "long"
	TypeDouble       Type = "double"
	TypeDatetime     Type = "datetime"
	TypeBoolean      Type = "boolean"
	TypeSparseVector Type = "sparseVector"
)

func (e Type) ToPointer() *Type {
	return &e
}

// IsExact returns true if the value matches a known enum value, false otherwise.
func (e *Type) IsExact() bool {
	if e != nil {
		switch *e {
		case "keyword", "long", "double", "datetime", "boolean", "sparseVector":
			return true
		}
	}
	return false
}

// IndexConfigs - Types that do not need additional parameters.
type IndexConfigs struct {
	Type Type `json:"type"`
}

func (i IndexConfigs) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(i, "", false)
}

func (i *IndexConfigs) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &i, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (i *IndexConfigs) GetType() Type {
	if i == nil {
		return Type("")
	}
	return i.Type
}

type TypeVector string

const (
	TypeVectorVector TypeVector = "vector"
)

func (e TypeVector) ToPointer() *TypeVector {
	return &e
}
func (e *TypeVector) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "vector":
		*e = TypeVector(v)
		return nil
	default:
		return fmt.Errorf("invalid value for TypeVector: %v", v)
	}
}

type EmbeddingConfigProvider string

const (
	EmbeddingConfigProviderOpenai EmbeddingConfigProvider = "openai"
)

func (e EmbeddingConfigProvider) ToPointer() *EmbeddingConfigProvider {
	return &e
}

// IsExact returns true if the value matches a known enum value, false otherwise.
func (e *EmbeddingConfigProvider) IsExact() bool {
	if e != nil {
		switch *e {
		case "openai":
			return true
		}
	}
	return false
}

// Similarity - Resolved vector similarity metric.
type Similarity string

const (
	SimilarityCosine          Similarity = "cosine"
	SimilarityEuclidean       Similarity = "euclidean"
	SimilarityDotProduct      Similarity = "dot_product"
	SimilarityMaxInnerProduct Similarity = "max_inner_product"
)

func (e Similarity) ToPointer() *Similarity {
	return &e
}

// IsExact returns true if the value matches a known enum value, false otherwise.
func (e *Similarity) IsExact() bool {
	if e != nil {
		switch *e {
		case "cosine", "euclidean", "dot_product", "max_inner_product":
			return true
		}
	}
	return false
}

// EmbeddingConfig - Managed embedding configuration for vector fields.
type EmbeddingConfig struct {
	// Embedding provider.
	Provider EmbeddingConfigProvider `json:"provider"`
	// Embedding model name.
	Model string `json:"model"`
	// Source text field name used to generate embeddings.
	SourceField string `json:"sourceField"`
	// Resolved embedding dimensions. Optional in requests and resolved in stored collection metadata.
	Dimensions *int64 `json:"dimensions,omitzero"`
	// Resolved vector similarity metric. Optional in requests and resolved in stored collection metadata.
	Similarity *Similarity `default:"cosine" json:"similarity,omitzero"`
}

func (e EmbeddingConfig) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(e, "", false)
}

func (e *EmbeddingConfig) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &e, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (e *EmbeddingConfig) GetProvider() EmbeddingConfigProvider {
	if e == nil {
		return EmbeddingConfigProvider("")
	}
	return e.Provider
}

func (e *EmbeddingConfig) GetModel() string {
	if e == nil {
		return ""
	}
	return e.Model
}

func (e *EmbeddingConfig) GetSourceField() string {
	if e == nil {
		return ""
	}
	return e.SourceField
}

func (e *EmbeddingConfig) GetDimensions() *int64 {
	if e == nil {
		return nil
	}
	return e.Dimensions
}

func (e *EmbeddingConfig) GetSimilarity() *Similarity {
	if e == nil {
		return nil
	}
	return e.Similarity
}

type IndexConfigsVector struct {
	Type TypeVector `json:"type"`
	// Set to false or omit for unmanaged vector fields.
	ManagedEmbedding *bool `json:"managedEmbedding,omitzero"`
	// Vector dimensions for unmanaged vector fields.
	Dimensions int64 `json:"dimensions"`
	// Vector similarity metric for unmanaged vector fields.
	Similarity *Similarity `default:"cosine" json:"similarity"`
}

func (i IndexConfigsVector) MarshalJSON() ([]byte, error) {
	if i.ManagedEmbedding != nil && *i.ManagedEmbedding {
		return nil, errors.New("managedEmbedding=true requires IndexConfigsManagedEmbeddingVector")
	}
	if i.Dimensions == 0 {
		return nil, errors.New("Dimensions is required field")
	}
	if i.Similarity != nil && !i.Similarity.IsExact() {
		return nil, fmt.Errorf("%s is not a supported similarity metric supported similarity metrics are %v", *i.Similarity, supportedVectorSimilarities())
	}
	return utils.MarshalJSON(i, "", false)
}

func (i *IndexConfigsVector) UnmarshalJSON(data []byte) error {
	state, err := vectorEmbeddingState(data)
	if err != nil {
		return err
	}
	if state.hasEmbedding {
		return errors.New("embedding is not allowed when managedEmbedding=false")
	}
	if state.hasManagedEmbedding && state.managedEmbedding {
		return errors.New("managedEmbedding=true requires IndexConfigsManagedEmbeddingVector")
	}
	if !state.hasDimensions {
		return errors.New("Dimensions is required field")
	}
	if err := utils.UnmarshalJSON(data, &i, "", false, nil); err != nil {
		return err
	}
	if i.Dimensions == 0 {
		return errors.New("Dimensions is required field")
	}
	if i.Similarity != nil && !i.Similarity.IsExact() {
		return fmt.Errorf("%s is not a supported similarity metric supported similarity metrics are %v", *i.Similarity, supportedVectorSimilarities())
	}
	return nil
}

func (i *IndexConfigsVector) GetType() TypeVector {
	if i == nil {
		return TypeVector("")
	}
	return i.Type
}

func (i *IndexConfigsVector) GetManagedEmbedding() *bool {
	if i == nil {
		return nil
	}
	return i.ManagedEmbedding
}

func (i *IndexConfigsVector) GetDimensions() int64 {
	if i == nil {
		return 0
	}
	return i.Dimensions
}

func (i *IndexConfigsVector) GetSimilarity() *Similarity {
	if i == nil {
		return nil
	}
	return i.Similarity
}

type IndexConfigsManagedEmbeddingVector struct {
	Type TypeVector `json:"type"`
	// Managed embedding vector field.
	ManagedEmbedding bool            `json:"managedEmbedding"`
	Embedding        EmbeddingConfig `json:"embedding"`
}

func (i IndexConfigsManagedEmbeddingVector) MarshalJSON() ([]byte, error) {
	i.ManagedEmbedding = true
	return utils.MarshalJSON(i, "", false)
}

func (i *IndexConfigsManagedEmbeddingVector) UnmarshalJSON(data []byte) error {
	state, err := vectorEmbeddingState(data)
	if err != nil {
		return err
	}
	if state.hasDimensions {
		return errors.New("Top-level dimensions are not allowed for managed embedding field")
	}
	if state.hasSimilarity {
		return errors.New("Top-level similarity is not allowed for managed embedding field")
	}
	if !state.hasManagedEmbedding || !state.managedEmbedding {
		if state.hasEmbedding {
			return errors.New("managedEmbedding=true is required when embedding config is provided")
		}
		return errors.New("managedEmbedding=true is required for managed embedding vector")
	}
	if !state.hasEmbedding {
		return errors.New("embedding is required when managedEmbedding=true")
	}
	if err := utils.UnmarshalJSON(data, &i, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (i *IndexConfigsManagedEmbeddingVector) GetType() TypeVector {
	if i == nil {
		return TypeVector("")
	}
	return i.Type
}

func (i *IndexConfigsManagedEmbeddingVector) GetManagedEmbedding() bool {
	if i == nil {
		return false
	}
	return i.ManagedEmbedding
}

func (i *IndexConfigsManagedEmbeddingVector) GetEmbedding() EmbeddingConfig {
	if i == nil {
		return EmbeddingConfig{}
	}
	return i.Embedding
}

type TypeText string

const (
	TypeTextText TypeText = "text"
)

func (e TypeText) ToPointer() *TypeText {
	return &e
}
func (e *TypeText) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "text":
		*e = TypeText(v)
		return nil
	default:
		return fmt.Errorf("invalid value for TypeText: %v", v)
	}
}

type Analyzer string

const (
	AnalyzerStandard Analyzer = "standard"
	AnalyzerKorean   Analyzer = "korean"
	AnalyzerJapanese Analyzer = "japanese"
	AnalyzerEnglish  Analyzer = "english"
)

func (e Analyzer) ToPointer() *Analyzer {
	return &e
}

// IsExact returns true if the value matches a known enum value, false otherwise.
func (e *Analyzer) IsExact() bool {
	if e != nil {
		switch *e {
		case "standard", "korean", "japanese", "english":
			return true
		}
	}
	return false
}

type IndexConfigsText struct {
	Type TypeText `json:"type"`
	// Analyzers.
	Analyzers []Analyzer `json:"analyzers,omitzero"`
}

func (i IndexConfigsText) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(i, "", false)
}

func (i *IndexConfigsText) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &i, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (i *IndexConfigsText) GetType() TypeText {
	if i == nil {
		return TypeText("")
	}
	return i.Type
}

func (i *IndexConfigsText) GetAnalyzers() []Analyzer {
	if i == nil {
		return nil
	}
	return i.Analyzers
}

type IndexConfigsUnionType string

const (
	IndexConfigsUnionTypeText         IndexConfigsUnionType = "text"
	IndexConfigsUnionTypeVector       IndexConfigsUnionType = "vector"
	IndexConfigsUnionTypeKeyword      IndexConfigsUnionType = "keyword"
	IndexConfigsUnionTypeLong         IndexConfigsUnionType = "long"
	IndexConfigsUnionTypeDouble       IndexConfigsUnionType = "double"
	IndexConfigsUnionTypeDatetime     IndexConfigsUnionType = "datetime"
	IndexConfigsUnionTypeBoolean      IndexConfigsUnionType = "boolean"
	IndexConfigsUnionTypeSparseVector IndexConfigsUnionType = "sparseVector"
	IndexConfigsUnionTypeObject       IndexConfigsUnionType = "object"
)

type IndexConfigsUnion struct {
	IndexConfigsText                   *IndexConfigsText                   `queryParam:"inline" union:"member"`
	IndexConfigsVector                 *IndexConfigsVector                 `queryParam:"inline" union:"member"`
	IndexConfigsManagedEmbeddingVector *IndexConfigsManagedEmbeddingVector `queryParam:"inline" union:"member"`
	IndexConfigs                       *IndexConfigs                       `queryParam:"inline" union:"member"`
	IndexConfigsObject                 *IndexConfigsObject                 `queryParam:"inline" union:"member"`

	Type IndexConfigsUnionType
}

func CreateIndexConfigsUnionText(text IndexConfigsText) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeText

	typStr := TypeText(typ)
	text.Type = typStr

	return IndexConfigsUnion{
		IndexConfigsText: &text,
		Type:             typ,
	}
}

func CreateIndexConfigsUnionVector(vector IndexConfigsVector) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeVector

	typStr := TypeVector(typ)
	vector.Type = typStr

	return IndexConfigsUnion{
		IndexConfigsVector: &vector,
		Type:               typ,
	}
}

func CreateIndexConfigsUnionManagedEmbeddingVector(managedEmbeddingVector IndexConfigsManagedEmbeddingVector) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeVector

	typStr := TypeVector(typ)
	managedEmbeddingVector.Type = typStr
	managedEmbeddingVector.ManagedEmbedding = true

	return IndexConfigsUnion{
		IndexConfigsManagedEmbeddingVector: &managedEmbeddingVector,
		Type:                               typ,
	}
}

func CreateIndexConfigsUnionKeyword(keyword IndexConfigs) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeKeyword

	typStr := Type(typ)
	keyword.Type = typStr

	return IndexConfigsUnion{
		IndexConfigs: &keyword,
		Type:         typ,
	}
}

func CreateIndexConfigsUnionLong(long IndexConfigs) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeLong

	typStr := Type(typ)
	long.Type = typStr

	return IndexConfigsUnion{
		IndexConfigs: &long,
		Type:         typ,
	}
}

func CreateIndexConfigsUnionDouble(double IndexConfigs) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeDouble

	typStr := Type(typ)
	double.Type = typStr

	return IndexConfigsUnion{
		IndexConfigs: &double,
		Type:         typ,
	}
}

func CreateIndexConfigsUnionDatetime(datetime IndexConfigs) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeDatetime

	typStr := Type(typ)
	datetime.Type = typStr

	return IndexConfigsUnion{
		IndexConfigs: &datetime,
		Type:         typ,
	}
}

func CreateIndexConfigsUnionBoolean(boolean IndexConfigs) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeBoolean

	typStr := Type(typ)
	boolean.Type = typStr

	return IndexConfigsUnion{
		IndexConfigs: &boolean,
		Type:         typ,
	}
}

func CreateIndexConfigsUnionSparseVector(sparseVector IndexConfigs) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeSparseVector

	typStr := Type(typ)
	sparseVector.Type = typStr

	return IndexConfigsUnion{
		IndexConfigs: &sparseVector,
		Type:         typ,
	}
}

func CreateIndexConfigsUnionObject(object IndexConfigsObject) IndexConfigsUnion {
	typ := IndexConfigsUnionTypeObject

	typStr := TypeObject(typ)
	object.Type = typStr

	return IndexConfigsUnion{
		IndexConfigsObject: &object,
		Type:               typ,
	}
}

func (u *IndexConfigsUnion) UnmarshalJSON(data []byte) error {

	type discriminator struct {
		Type string `json:"type"`
	}

	dis := new(discriminator)
	if err := json.Unmarshal(data, &dis); err != nil {
		return fmt.Errorf("could not unmarshal discriminator: %w", err)
	}

	switch dis.Type {
	case "text":
		indexConfigsText := new(IndexConfigsText)
		if err := utils.UnmarshalJSON(data, &indexConfigsText, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == text) type IndexConfigsText within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigsText = indexConfigsText
		u.Type = IndexConfigsUnionTypeText
		return nil
	case "vector":
		state, err := vectorEmbeddingState(data)
		if err != nil {
			return fmt.Errorf("could not inspect vector embedding state within IndexConfigsUnion: %w", err)
		}
		if state.hasEmbedding && !state.hasManagedEmbedding {
			return errors.New("managedEmbedding=true is required when embedding config is provided")
		}
		if state.hasEmbedding && !state.managedEmbedding {
			return errors.New("embedding is not allowed when managedEmbedding=false")
		}
		if state.hasManagedEmbedding && state.managedEmbedding {
			indexConfigsManagedEmbeddingVector := new(IndexConfigsManagedEmbeddingVector)
			if err := utils.UnmarshalJSON(data, &indexConfigsManagedEmbeddingVector, "", true, nil); err != nil {
				return fmt.Errorf("could not unmarshal `%s` into expected (Type == vector, ManagedEmbedding == true) type IndexConfigsManagedEmbeddingVector within IndexConfigsUnion: %w", string(data), err)
			}

			u.IndexConfigsManagedEmbeddingVector = indexConfigsManagedEmbeddingVector
			u.Type = IndexConfigsUnionTypeVector
			return nil
		}

		indexConfigsVector := new(IndexConfigsVector)
		if err := utils.UnmarshalJSON(data, &indexConfigsVector, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == vector) type IndexConfigsVector within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigsVector = indexConfigsVector
		u.Type = IndexConfigsUnionTypeVector
		return nil
	case "keyword":
		indexConfigs := new(IndexConfigs)
		if err := utils.UnmarshalJSON(data, &indexConfigs, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == keyword) type IndexConfigs within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigs = indexConfigs
		u.Type = IndexConfigsUnionTypeKeyword
		return nil
	case "long":
		indexConfigs := new(IndexConfigs)
		if err := utils.UnmarshalJSON(data, &indexConfigs, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == long) type IndexConfigs within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigs = indexConfigs
		u.Type = IndexConfigsUnionTypeLong
		return nil
	case "double":
		indexConfigs := new(IndexConfigs)
		if err := utils.UnmarshalJSON(data, &indexConfigs, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == double) type IndexConfigs within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigs = indexConfigs
		u.Type = IndexConfigsUnionTypeDouble
		return nil
	case "datetime":
		indexConfigs := new(IndexConfigs)
		if err := utils.UnmarshalJSON(data, &indexConfigs, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == datetime) type IndexConfigs within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigs = indexConfigs
		u.Type = IndexConfigsUnionTypeDatetime
		return nil
	case "boolean":
		indexConfigs := new(IndexConfigs)
		if err := utils.UnmarshalJSON(data, &indexConfigs, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == boolean) type IndexConfigs within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigs = indexConfigs
		u.Type = IndexConfigsUnionTypeBoolean
		return nil
	case "sparseVector":
		indexConfigs := new(IndexConfigs)
		if err := utils.UnmarshalJSON(data, &indexConfigs, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == sparseVector) type IndexConfigs within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigs = indexConfigs
		u.Type = IndexConfigsUnionTypeSparseVector
		return nil
	case "object":
		indexConfigsObject := new(IndexConfigsObject)
		if err := utils.UnmarshalJSON(data, &indexConfigsObject, "", true, nil); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == object) type IndexConfigsObject within IndexConfigsUnion: %w", string(data), err)
		}

		u.IndexConfigsObject = indexConfigsObject
		u.Type = IndexConfigsUnionTypeObject
		return nil
	}

	return fmt.Errorf("could not unmarshal `%s` into any supported union types for IndexConfigsUnion", string(data))
}

func (u IndexConfigsUnion) MarshalJSON() ([]byte, error) {
	if u.IndexConfigsText != nil {
		return utils.MarshalJSON(u.IndexConfigsText, "", true)
	}

	if u.IndexConfigsVector != nil {
		return utils.MarshalJSON(u.IndexConfigsVector, "", true)
	}

	if u.IndexConfigsManagedEmbeddingVector != nil {
		return utils.MarshalJSON(u.IndexConfigsManagedEmbeddingVector, "", true)
	}

	if u.IndexConfigs != nil {
		return utils.MarshalJSON(u.IndexConfigs, "", true)
	}

	if u.IndexConfigsObject != nil {
		return utils.MarshalJSON(u.IndexConfigsObject, "", true)
	}

	return nil, errors.New("could not marshal union type IndexConfigsUnion: all fields are null")
}

type vectorEmbeddingRawState struct {
	hasManagedEmbedding bool
	managedEmbedding    bool
	hasEmbedding        bool
	hasDimensions       bool
	hasSimilarity       bool
}

func vectorEmbeddingState(data []byte) (vectorEmbeddingRawState, error) {
	state := vectorEmbeddingRawState{}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return state, err
	}

	if raw, ok := fields["managedEmbedding"]; ok && !isJSONNull(raw) {
		state.hasManagedEmbedding = true
		if err := json.Unmarshal(raw, &state.managedEmbedding); err != nil {
			return state, fmt.Errorf("could not unmarshal managedEmbedding: %w", err)
		}
	}
	if raw, ok := fields["embedding"]; ok && !isJSONNull(raw) {
		state.hasEmbedding = true
	}
	if raw, ok := fields["dimensions"]; ok && !isJSONNull(raw) {
		state.hasDimensions = true
	}
	if raw, ok := fields["similarity"]; ok && !isJSONNull(raw) {
		state.hasSimilarity = true
	}
	return state, nil
}

func isJSONNull(raw json.RawMessage) bool {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return false
	}
	return v == nil
}

func supportedVectorSimilarities() []Similarity {
	return []Similarity{
		SimilarityCosine,
		SimilarityEuclidean,
		SimilarityDotProduct,
		SimilarityMaxInnerProduct,
	}
}
