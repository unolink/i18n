[English version (README.md)](README.md)

# i18n

Легковесная интернационализация для Go с YAML-переводами, фоллбэком локалей и передачей через контекст.

## Возможности

- **Любая локаль** — `Locale` это строковый тип, не ограничен фиксированным набором
- **YAML-переводы** — плоские файлы ключ-значение, по одному на локаль
- **Фоллбэк локали** — нет перевода на `de`? Используется настроенная локаль по умолчанию
- **Передача через контекст** — per-request локаль через `context.Context`, безопасно для конкурентных веб-серверов
- **fmt.Sprintf форматирование** — `"Привет, %s!"` с аргументами
- **Несколько источников** — `embed.FS`, файлы на диске, сырые байты
- **Глобальный или per-instance** — пакетная функция `T()` для удобства, или свой `*Translator`
- **Потокобезопасность** — конкурентное чтение из любого числа горутин
- **Одна зависимость** — только `gopkg.in/yaml.v3`

## Установка

```bash
go get github.com/unolink/i18n
```

## Быстрый старт

Создай файлы переводов:

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

Загрузи и используй:

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

## Фоллбэк локали

Каждый транслятор имеет фоллбэк-локаль, заданную при создании. Если ключ не найден в запрошенной локали, пробуется фоллбэк. Если и там нет — возвращается сам ключ.

```go
t := i18n.NewTranslator(i18n.EN) // EN — фоллбэк
t.LoadFromBytes(i18n.EN, []byte(`greeting: Hello`))

t.Translate("de", "greeting") // "Hello" (фоллбэк на EN)
t.Translate("de", "missing")  // "missing" (ключ как есть)
```

## Передача через контекст

Для веб-серверов — привязка локали к контексту запроса:

```go
func localeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        locale := detectLocale(r) // ваша логика
        ctx := i18n.WithLocale(r.Context(), i18n.Locale(locale))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    msg := i18n.TCtx(r.Context(), "greeting") // берёт локаль из контекста
    w.Write([]byte(msg))
}
```

## Произвольные локали

`Locale` — обычная строка. Используй любой код локали:

```go
t := i18n.NewTranslator("de")
t.LoadFromBytes("de", deYAML)
t.LoadFromBytes("fr", frYAML)
t.LoadFromBytes("ja", jaYAML)

t.Translate("ja", "greeting") // японский
t.Translate("fr", "greeting") // французский
t.T("greeting")               // немецкий (фоллбэк)
```

## Загрузка переводов

### Из embed.FS

```go
//go:embed locales/*.yaml
var localesFS embed.FS

t.LoadFromFS(localesFS, "locales") // локаль извлекается из имени файла
```

### С диска

```go
t.LoadFromFile(i18n.EN, "/etc/myapp/en.yaml")
t.LoadFromFile(i18n.RU, "/etc/myapp/ru.yaml")
```

### Из байтов

```go
t.LoadFromBytes("es", []byte(`greeting: Hola`))
```

## Обзор API

| Функция | Описание |
|---|---|
| `NewTranslator(fallback)` | Создать транслятор с фоллбэк-локалью |
| `SetDefault(t)` / `Default()` | Установить / получить глобальный транслятор |
| `t.LoadFromFS(fs, dir)` | Загрузить YAML из `fs.FS` |
| `t.LoadFromFile(locale, path)` | Загрузить один YAML-файл |
| `t.LoadFromBytes(locale, data)` | Загрузить переводы из байтов |
| `t.T(key, args...)` | Перевести с фоллбэк-локалью |
| `t.Translate(locale, key, args...)` | Перевести для конкретной локали |
| `t.TCtx(ctx, key, args...)` | Перевести с локалью из контекста |
| `t.LookupKey(locale, key)` | Поиск без fmt.Sprintf форматирования |
| `t.Has(key)` / `t.HasLocale(l, key)` | Проверить наличие перевода |
| `WithLocale(ctx, locale)` | Привязать локаль к контексту |
| `LocaleFromContext(ctx)` | Извлечь локаль из контекста |
| `T()`, `Translate()`, `TCtx()`, `LookupKey()` | Пакетные функции через глобальный транслятор |

## Зависимости

- `gopkg.in/yaml.v3` — парсинг YAML

## Лицензия

MIT — см. [LICENSE](LICENSE).
