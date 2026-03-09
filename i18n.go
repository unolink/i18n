// Package i18n provides lightweight internationalization for Go applications.
//
// Translations are stored in flat YAML files (one file per locale) and loaded
// at startup via [Translator.LoadFromEmbed], [Translator.LoadFromFile], or
// [Translator.LoadFromBytes]. The package does not embed any translations
// itself — the consumer supplies them.
//
// A global default translator can be set with [SetDefault] and used through
// package-level convenience functions [T], [TCtx], and [Translate].
// For per-request locale handling in web servers, use [WithLocale] and
// [LocaleFromContext] to propagate locale through context.
package i18n

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Locale represents a language/locale code (e.g., "ru", "en", "de").
type Locale string

// Common locale constants for convenience.
// Any string can be used as a Locale — these are not the only supported values.
const (
	EN Locale = "en"
	RU Locale = "ru"
)

// Translator provides thread-safe translation with locale fallback.
// Locale is NOT stored as global state — pass it per-request via context.
type Translator struct {
	dictionary *Dictionary
	fallback   Locale
}

// defaultTranslator is the global translator instance.
// Uses atomic.Pointer for lock-free concurrent reads.
var defaultTranslator atomic.Pointer[Translator]

// NewTranslator creates a new translator with the given fallback locale.
// The fallback locale is used when a translation is not found in the requested locale.
func NewTranslator(fallback Locale) *Translator {
	if fallback == "" {
		fallback = EN
	}
	return &Translator{
		dictionary: NewDictionary(),
		fallback:   fallback,
	}
}

// SetDefault sets the global default translator.
// Must be called before using package-level functions T, TCtx, Translate.
func SetDefault(t *Translator) {
	defaultTranslator.Store(t)
}

// Default returns the global default translator, or nil if not set.
func Default() *Translator {
	return defaultTranslator.Load()
}

// Fallback returns the fallback locale of the translator.
func (t *Translator) Fallback() Locale {
	return t.fallback
}

// T translates a key using the fallback locale with optional fmt.Sprintf arguments.
// If the translation is not found, returns the key itself.
func (t *Translator) T(key string, args ...any) string {
	return t.Translate(t.fallback, key, args...)
}

// Translate translates a key for a specific locale with optional fmt.Sprintf arguments.
// If the translation is not found for the requested locale, falls back to the
// translator's fallback locale. If still not found, returns the key itself.
func (t *Translator) Translate(locale Locale, key string, args ...any) string {
	// Try requested locale.
	if translation := t.dictionary.Get(locale, key); translation != "" {
		if len(args) > 0 {
			return fmt.Sprintf(translation, args...)
		}
		return translation
	}

	// Fallback to default locale.
	if locale != t.fallback {
		if translation := t.dictionary.Get(t.fallback, key); translation != "" {
			if len(args) > 0 {
				return fmt.Sprintf(translation, args...)
			}
			return translation
		}
	}

	// Return key as-is, formatted if args provided.
	if len(args) > 0 {
		return fmt.Sprintf(key, args...)
	}
	return key
}

// LookupKey returns the translation for the given key and locale,
// performing locale fallback without fmt.Sprintf processing.
// Use this when the translated value does not contain format directives
// (e.g., in HTML templates where values are interpolated by the template engine).
func (t *Translator) LookupKey(locale Locale, key string) string {
	if translation := t.dictionary.Get(locale, key); translation != "" {
		return translation
	}
	if locale != t.fallback {
		if translation := t.dictionary.Get(t.fallback, key); translation != "" {
			return translation
		}
	}
	return key
}

// Has reports whether a translation exists for the given key
// in the requested locale or the fallback locale.
func (t *Translator) Has(key string) bool {
	return t.dictionary.Has(t.fallback, key)
}

// HasLocale reports whether a translation exists for the given key in the given locale.
func (t *Translator) HasLocale(locale Locale, key string) bool {
	return t.dictionary.Has(locale, key)
}

// Dictionary returns the underlying dictionary for direct access.
func (t *Translator) Dictionary() *Dictionary {
	return t.dictionary
}

// Context key for storing locale in context.
type localeContextKey struct{}

// WithLocale returns a new context carrying the specified locale.
func WithLocale(ctx context.Context, locale Locale) context.Context {
	return context.WithValue(ctx, localeContextKey{}, locale)
}

// LocaleFromContext extracts locale from context.
// Returns the translator's fallback locale if not found in context,
// or EN if no default translator is set.
func LocaleFromContext(ctx context.Context) Locale {
	if locale, ok := ctx.Value(localeContextKey{}).(Locale); ok {
		return locale
	}
	if t := defaultTranslator.Load(); t != nil {
		return t.fallback
	}
	return EN
}

// TCtx translates a key using locale from context.
// Extracts locale via LocaleFromContext, then delegates to Translate.
func (t *Translator) TCtx(ctx context.Context, key string, args ...any) string {
	return t.Translate(LocaleFromContext(ctx), key, args...)
}

// --- Package-level convenience functions using the default translator ---

// T translates a key using the default translator's fallback locale.
// Returns the key itself if no default translator is set.
func T(key string, args ...any) string {
	t := defaultTranslator.Load()
	if t == nil {
		if len(args) > 0 {
			return fmt.Sprintf(key, args...)
		}
		return key
	}
	return t.T(key, args...)
}

// Translate translates a key for a specific locale using the default translator.
func Translate(locale Locale, key string, args ...any) string {
	t := defaultTranslator.Load()
	if t == nil {
		if len(args) > 0 {
			return fmt.Sprintf(key, args...)
		}
		return key
	}
	return t.Translate(locale, key, args...)
}

// TCtx translates a key using locale from context via the default translator.
func TCtx(ctx context.Context, key string, args ...any) string {
	t := defaultTranslator.Load()
	if t == nil {
		if len(args) > 0 {
			return fmt.Sprintf(key, args...)
		}
		return key
	}
	return t.TCtx(ctx, key, args...)
}

// LookupKey performs key lookup using the default translator without formatting.
func LookupKey(locale Locale, key string) string {
	t := defaultTranslator.Load()
	if t == nil {
		return key
	}
	return t.LookupKey(locale, key)
}
