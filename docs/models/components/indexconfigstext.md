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
