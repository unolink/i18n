[Русская версия (README.ru.md)](README.ru.md)

# i18n

Lightweight internationalization for Go with YAML translations, locale fallback, and context-aware lookups.

## Features

- **Any locale** — `Locale` is a string type; not limited to a fixed set
- **YAML translations** — flat key-value files, one per locale
- **Locale fallback** — missing translation in `de`? Falls back to the configured default
- **Context propagation** — per-request locale via `context.Context`, safe for concurrent web servers
- **fmt.Sprintf formatting** — `"Hello, %s!"` with arguments
- **Multiple load sources** — `embed.FS`, disk files, raw bytes
- **Global or per-instance** — package-level `T()` for convenience, or bring your own `*Translator`
- **Thread-safe** — concurrent reads from any number of goroutines
- **One dependency** — only `gopkg.in/yaml.v3`

## Install

```bash
go get github.com/unolink/i18n
```

## Quick Start

Create translation files:

```yaml
# locales/en.yaml
greeting: "Hello"
greeting.name: "Hello, %s!"
error.not_found: "Resource not found"
```

```yaml
# locales/ru.yaml
greeting: "Привет"
greeting.name: "Привет, %s!"
error.not_found: "Ресурс не найден"
```

Load and use:

```go
import "github.com/unolink/i18n"

//go:embed locales/*.yaml
var localesFS embed.FS

func main() {
    t := i18n.NewTranslator(i18n.EN)
    t.LoadFromFS(localesFS, "locales")
    i18n.SetDefault(t)

    fmt.Println(i18n.T("greeting"))                        // Hello
    fmt.Println(i18n.Translate(i18n.RU, "greeting"))       // Привет
    fmt.Println(i18n.Translate(i18n.EN, "greeting.name", "Alice")) // Hello, Alice!
}
```

## Locale Fallback

Each translator has a fallback locale set at creation. If a key is missing in the requested locale, the fallback is tried. If still missing, the key itself is returned.

```go
t := i18n.NewTranslator(i18n.EN) // EN is the fallback
t.LoadFromBytes(i18n.EN, []byte(`greeting: Hello`))

t.Translate("de", "greeting") // "Hello" (fallback to EN)
t.Translate("de", "missing")  // "missing" (key returned as-is)
```

## Context Propagation

For web servers, attach locale to the request context:

```go
func localeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        locale := detectLocale(r) // your logic
        ctx := i18n.WithLocale(r.Context(), i18n.Locale(locale))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    msg := i18n.TCtx(r.Context(), "greeting") // uses locale from context
    w.Write([]byte(msg))
}
```

## Custom Locales

`Locale` is a plain `string`. Use any locale code you need:

```go
t := i18n.NewTranslator("de")
t.LoadFromBytes("de", deYAML)
t.LoadFromBytes("fr", frYAML)
t.LoadFromBytes("ja", jaYAML)

t.Translate("ja", "greeting") // Japanese
t.Translate("fr", "greeting") // French
t.T("greeting")               // German (fallback)
```

## Loading Translations

### From embed.FS

```go
//go:embed locales/*.yaml
var localesFS embed.FS

t.LoadFromFS(localesFS, "locales") // locale extracted from filename
```

### From disk

```go
t.LoadFromFile(i18n.EN, "/etc/myapp/en.yaml")
t.LoadFromFile(i18n.RU, "/etc/myapp/ru.yaml")
```

### From bytes

```go
t.LoadFromBytes("es", []byte(`greeting: Hola`))
```

## API Summary

| Function | Description |
|---|---|
| `NewTranslator(fallback)` | Create a translator with a fallback locale |
| `SetDefault(t)` / `Default()` | Set / get the global translator |
| `t.LoadFromFS(fs, dir)` | Load YAML files from an `fs.FS` |
| `t.LoadFromFile(locale, path)` | Load a single YAML file |
| `t.LoadFromBytes(locale, data)` | Load translations from bytes |
| `t.T(key, args...)` | Translate using fallback locale |
| `t.Translate(locale, key, args...)` | Translate for a specific locale |
| `t.TCtx(ctx, key, args...)` | Translate using locale from context |
| `t.LookupKey(locale, key)` | Lookup without fmt.Sprintf formatting |
| `t.Has(key)` / `t.HasLocale(l, key)` | Check if translation exists |
| `WithLocale(ctx, locale)` | Attach locale to context |
| `LocaleFromContext(ctx)` | Extract locale from context |
| `T()`, `Translate()`, `TCtx()`, `LookupKey()` | Package-level shortcuts using default translator |

## Dependencies

- `gopkg.in/yaml.v3` — YAML parsing

## License

MIT — see [LICENSE](LICENSE).
