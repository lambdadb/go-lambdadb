# IndexConfigsText


## Fields

| Field                                                        | Type                                                         | Required                                                     | Description                                                  |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| `Type`                                                       | [components.TypeText](../../models/components/typetext.md)   | :heavy_check_mark:                                           | N/A                                                          |
| `Analyzers`                                                  | [][components.Analyzer](../../models/components/analyzer.md) | :heavy_minus_sign:                                           | Text analyzers applied independently to this field.                                                   |

Leave `Analyzers` nil to omit `analyzers` and use the server default
`["standard"]`. An explicit empty slice sends `[]` and does not select the
default. The SDK preserves names, order, and duplicates without local
validation; unknown strings can be represented as `components.Analyzer("name")`.
Use the documented lowercase names without duplicates for new configurations.
The server validation in [PR #422](https://github.com/lambdadb/lambdadb/pull/422)
(`e208b3327a49ade2e429bdcfcd006582a144eee3`) rejects duplicate analyzer names
case-insensitively with HTTP 400. Preserving a value in SDK JSON does not imply
that the server accepts it.

For example, this text field selects Chinese word segmentation:

```go
components.CreateIndexConfigsUnionText(components.IndexConfigsText{
    Type: components.TypeTextText,
    Analyzers: []components.Analyzer{components.AnalyzerChinese},
})
```

`AnalyzerChinese` uses SmartChinese word segmentation; `AnalyzerCjk` uses
character bigrams. To evaluate both, pass
`[]components.Analyzer{components.AnalyzerChinese, components.AnalyzerCjk}`.
Every selected analyzer indexes the field separately; combining them does not
provide fallback or automatic language detection and can increase indexing,
storage, and query work. Existing fields cannot change their analyzer list;
create a new field or collection and reindex the source text.

The added constants require a server environment that supports them. Their
presence in the SDK does not establish service availability.

All [49 analyzer names](analyzer.md) select fixed presets with their default
settings; this SDK does not configure custom pipelines or analyzer options.
For example, select the keyword analyzer on a **text** field:

```go
components.CreateIndexConfigsUnionText(components.IndexConfigsText{
    Type: components.TypeTextText,
    Analyzers: []components.Analyzer{components.AnalyzerKeyword},
})
```

This remains a text field and does not enable keyword-field sorting or facets.
`AnalyzerNepali`, `AnalyzerTamil`, and `AnalyzerTelugu` are Lucene extensions;
they are not documented as shared Elasticsearch/OpenSearch presets.
