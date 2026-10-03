# Analyzer

`Analyzer` is an extensible string type. `IsExact()` reports whether a value
matches one of the known lowercase names below; it returns false for unknown
strings, case variants, and a nil pointer. It does not validate or reject values
during JSON serialization or deserialization. Server-side acceptance depends on
the target environment.

All 49 known names are fixed presets for `text` fields, with each preset's
server default configuration. Custom analyzer pipelines and options are not
exposed. `AnalyzerKeyword` remains a text analyzer; it does not change a field
to the `keyword` type or add that type's sorting/facet behavior.
`nepali`, `tamil`, and `telugu` are Lucene extensions, not common
Elasticsearch/OpenSearch support.

The expanded list follows [backend PR #437](https://github.com/lambdadb/lambdadb/pull/437),
merged as [55d888299fee44466326a9db8016af9811ade13b](https://github.com/lambdadb/lambdadb/blob/55d888299fee44466326a9db8016af9811ade13b/core/src/main/java/ai/lambdadb/core/IndexingConstants.java).
The upstream [OpenAPI at 961561c379acb079aec20191e13b89809ef096e9](https://github.com/lambdadb/docs/blob/961561c379acb079aec20191e13b89809ef096e9/reference/api/openapi.json)
still lists the original 16 names; its wire shape and default remain unchanged.
These are source contracts, not deployment evidence.

## Values

| Name | Value |
| --- | --- |
| `AnalyzerStandard` | standard |
| `AnalyzerKorean` | korean |
| `AnalyzerJapanese` | japanese |
| `AnalyzerEnglish` | english |
| `AnalyzerChinese` | chinese |
| `AnalyzerCjk` | cjk |
| `AnalyzerArabic` | arabic |
| `AnalyzerFrench` | french |
| `AnalyzerGerman` | german |
| `AnalyzerHindi` | hindi |
| `AnalyzerIndonesian` | indonesian |
| `AnalyzerItalian` | italian |
| `AnalyzerPortuguese` | portuguese |
| `AnalyzerRussian` | russian |
| `AnalyzerSpanish` | spanish |
| `AnalyzerTurkish` | turkish |
| `AnalyzerArmenian` | armenian |
| `AnalyzerBasque` | basque |
| `AnalyzerBengali` | bengali |
| `AnalyzerBrazilian` | brazilian |
| `AnalyzerBulgarian` | bulgarian |
| `AnalyzerCatalan` | catalan |
| `AnalyzerCzech` | czech |
| `AnalyzerDanish` | danish |
| `AnalyzerDutch` | dutch |
| `AnalyzerEstonian` | estonian |
| `AnalyzerFinnish` | finnish |
| `AnalyzerGalician` | galician |
| `AnalyzerGreek` | greek |
| `AnalyzerHungarian` | hungarian |
| `AnalyzerIrish` | irish |
| `AnalyzerLatvian` | latvian |
| `AnalyzerLithuanian` | lithuanian |
| `AnalyzerNorwegian` | norwegian |
| `AnalyzerPersian` | persian |
| `AnalyzerRomanian` | romanian |
| `AnalyzerSerbian` | serbian |
| `AnalyzerSorani` | sorani |
| `AnalyzerSwedish` | swedish |
| `AnalyzerThai` | thai |
| `AnalyzerSimple` | simple |
| `AnalyzerWhitespace` | whitespace |
| `AnalyzerStop` | stop |
| `AnalyzerKeyword` | keyword |
| `AnalyzerPattern` | pattern |
| `AnalyzerFingerprint` | fingerprint |
| `AnalyzerNepali` | nepali |
| `AnalyzerTamil` | tamil |
| `AnalyzerTelugu` | telugu |

See [IndexConfigsText](indexconfigstext.md) for configuration and omission behavior.
