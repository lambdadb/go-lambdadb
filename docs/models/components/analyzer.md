# Analyzer

`Analyzer` is an extensible string type. `IsExact()` reports whether a value
matches one of the known lowercase names below; it returns false for unknown
strings, case variants, and a nil pointer. It does not validate or reject values
during JSON serialization or deserialization. Server-side acceptance depends on
the target environment.

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

See [IndexConfigsText](indexconfigstext.md) for configuration and omission behavior.
