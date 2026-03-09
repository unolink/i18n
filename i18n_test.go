package i18n

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
)

// newTestTranslator creates a translator preloaded with test translations.
func newTestTranslator(t *testing.T) *Translator {
	t.Helper()
	tr := NewTranslator(EN)

	tr.Dictionary().SetBatch(EN, map[string]string{
		"greeting":        "Hello",
		"greeting.name":   "Hello, %s!",
		"item.count":      "%d item(s) found",
		"error.not_found": "Resource not found",
		"multi.args":      "%s has %d messages",
	})
	tr.Dictionary().SetBatch(RU, map[string]string{
		"greeting":        "Привет",
		"greeting.name":   "Привет, %s!",
		"item.count":      "Найдено элементов: %d",
		"error.not_found": "Ресурс не найден",
		"multi.args":      "У %s %d сообщений",
	})

	return tr
}

func TestTranslate_Basic(t *testing.T) {
	t.Parallel()
	tr := newTestTranslator(t)

	tests := []struct {
		name     string
		locale   Locale
		key      string
		expected string
	}{
		{"english", EN, "greeting", "Hello"},
		{"russian", RU, "greeting", "Привет"},
		{"english not_found", EN, "error.not_found", "Resource not found"},
		{"russian not_found", RU, "error.not_found", "Ресурс не найден"},
		{"missing key returns key", EN, "no.such.key", "no.such.key"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tr.Translate(tt.locale, tt.key)
			if got != tt.expected {
				t.Errorf("Translate(%q, %q) = %q, want %q", tt.locale, tt.key, got, tt.expected)
			}
		})
	}
}

func TestTranslate_Formatting(t *testing.T) {
	t.Parallel()
	tr := newTestTranslator(t)

	tests := []struct {
		name     string
		locale   Locale
		key      string
		args     []any
		expected string
	}{
		{"english name", EN, "greeting.name", []any{"Alice"}, "Hello, Alice!"},
		{"russian name", RU, "greeting.name", []any{"Alice"}, "Привет, Alice!"},
		{"english count", EN, "item.count", []any{42}, "42 item(s) found"},
		{"russian count", RU, "item.count", []any{42}, "Найдено элементов: 42"},
		{"multi args", EN, "multi.args", []any{"Bob", 5}, "Bob has 5 messages"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tr.Translate(tt.locale, tt.key, tt.args...)
			if got != tt.expected {
				t.Errorf("Translate(%q, %q, %v) = %q, want %q", tt.locale, tt.key, tt.args, got, tt.expected)
			}
		})
	}
}

func TestTranslate_Fallback(t *testing.T) {
	t.Parallel()
	tr := NewTranslator(EN)

	// Only English has this key.
	tr.Dictionary().Set(EN, "only.english", "English only value")

	// Requesting German (not loaded) should fall back to EN.
	got := tr.Translate("de", "only.english")
	if got != "English only value" {
		t.Errorf("expected fallback to EN, got %q", got)
	}

	// Key missing in all locales returns the key.
	got = tr.Translate("de", "no.such.key")
	if got != "no.such.key" {
		t.Errorf("expected key itself, got %q", got)
	}
}

func TestT_DefaultLocale(t *testing.T) {
	t.Parallel()
	tr := NewTranslator(RU)
	tr.Dictionary().Set(RU, "msg", "Сообщение")
	tr.Dictionary().Set(EN, "msg", "Message")

	// T() uses the fallback locale (RU).
	got := tr.T("msg")
	if got != "Сообщение" {
		t.Errorf("T(msg) = %q, want Сообщение", got)
	}
}

func TestLookupKey_NoFormatting(t *testing.T) {
	t.Parallel()
	tr := NewTranslator(EN)
	tr.Dictionary().Set(EN, "tpl", "Hello, %s!")

	// LookupKey must NOT apply Sprintf.
	got := tr.LookupKey(EN, "tpl")
	if got != "Hello, %s!" {
		t.Errorf("LookupKey should not format, got %q", got)
	}
}

func TestHas(t *testing.T) {
	t.Parallel()
	tr := newTestTranslator(t)

	if !tr.Has("greeting") {
		t.Error("Has(greeting) = false, want true")
	}
	if tr.Has("no.such.key") {
		t.Error("Has(no.such.key) = true, want false")
	}
}

func TestHasLocale(t *testing.T) {
	t.Parallel()
	tr := NewTranslator(EN)
	tr.Dictionary().Set(EN, "en.only", "value")

	if !tr.HasLocale(EN, "en.only") {
		t.Error("HasLocale(EN, en.only) = false, want true")
	}
	if tr.HasLocale(RU, "en.only") {
		t.Error("HasLocale(RU, en.only) = true, want false")
	}
}

func TestContext(t *testing.T) {
	t.Parallel()
	tr := newTestTranslator(t)

	ctx := WithLocale(context.Background(), RU)
	got := tr.TCtx(ctx, "greeting")
	if got != "Привет" {
		t.Errorf("TCtx with RU context = %q, want Привет", got)
	}

	// Without locale in context — fallback.
	got = tr.TCtx(context.Background(), "greeting")
	if got != "Hello" {
		t.Errorf("TCtx without locale context = %q, want Hello", got)
	}
}

func TestLocaleFromContext_NoDefault(t *testing.T) {
	t.Parallel()

	// Temporarily clear default translator.
	old := defaultTranslator.Load()
	defaultTranslator.Store(nil)
	defer defaultTranslator.Store(old)

	locale := LocaleFromContext(context.Background())
	if locale != EN {
		t.Errorf("LocaleFromContext without default = %q, want EN", locale)
	}
}

func TestSetDefault_PackageFunctions(t *testing.T) {
	tr := newTestTranslator(t)
	SetDefault(tr)
	defer SetDefault(nil)

	if got := T("greeting"); got != "Hello" {
		t.Errorf("T(greeting) = %q, want Hello", got)
	}

	if got := Translate(RU, "greeting"); got != "Привет" {
		t.Errorf("Translate(RU, greeting) = %q, want Привет", got)
	}

	ctx := WithLocale(context.Background(), RU)
	if got := TCtx(ctx, "greeting"); got != "Привет" {
		t.Errorf("TCtx(RU, greeting) = %q, want Привет", got)
	}

	if got := LookupKey(EN, "greeting"); got != "Hello" {
		t.Errorf("LookupKey(EN, greeting) = %q, want Hello", got)
	}
}

func TestPackageFunctions_NoDefault(t *testing.T) {
	old := defaultTranslator.Load()
	defaultTranslator.Store(nil)
	defer defaultTranslator.Store(old)

	if got := T("some.key"); got != "some.key" {
		t.Errorf("T without default = %q, want some.key", got)
	}
	if got := T("hello %s", "world"); got != "hello world" {
		t.Errorf("T with args without default = %q, want 'hello world'", got)
	}
	if got := LookupKey(EN, "some.key"); got != "some.key" {
		t.Errorf("LookupKey without default = %q, want some.key", got)
	}
}

func TestCustomLocale(t *testing.T) {
	t.Parallel()
	tr := NewTranslator("de")

	tr.Dictionary().SetBatch("de", map[string]string{
		"greeting": "Hallo",
	})
	tr.Dictionary().SetBatch("fr", map[string]string{
		"greeting": "Bonjour",
	})

	if got := tr.T("greeting"); got != "Hallo" {
		t.Errorf("T(greeting) with de fallback = %q, want Hallo", got)
	}
	if got := tr.Translate("fr", "greeting"); got != "Bonjour" {
		t.Errorf("Translate(fr, greeting) = %q, want Bonjour", got)
	}
}

func TestDictionary_Operations(t *testing.T) {
	t.Parallel()
	d := NewDictionary()

	d.Set(EN, "key1", "value1")
	if !d.Has(EN, "key1") {
		t.Error("Has(key1) = false after Set")
	}
	if got := d.Get(EN, "key1"); got != "value1" {
		t.Errorf("Get(key1) = %q, want value1", got)
	}

	// Missing key.
	if d.Has(EN, "missing") {
		t.Error("Has(missing) = true")
	}
	if got := d.Get(EN, "missing"); got != "" {
		t.Errorf("Get(missing) = %q, want empty", got)
	}

	// Batch.
	d.SetBatch(RU, map[string]string{"a": "1", "b": "2"})
	if got := d.Get(RU, "a"); got != "1" {
		t.Errorf("batch Get(a) = %q, want 1", got)
	}

	// Locale.
	all := d.Locale(RU)
	if len(all) != 2 {
		t.Errorf("Locale(RU) len = %d, want 2", len(all))
	}

	// Locales.
	locales := d.Locales()
	if len(locales) < 2 {
		t.Errorf("Locales() len = %d, want >= 2", len(locales))
	}

	// Clear.
	d.Clear(RU)
	if d.Has(RU, "a") {
		t.Error("Has(a) = true after Clear(RU)")
	}

	// ClearAll.
	d.ClearAll()
	if d.Has(EN, "key1") {
		t.Error("Has(key1) = true after ClearAll")
	}
}

func TestLoadFromFS(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"locales/en.yaml": &fstest.MapFile{
			Data: []byte("greeting: Hello\nfarewell: Goodbye\n"),
		},
		"locales/ru.yaml": &fstest.MapFile{
			Data: []byte("greeting: Привет\nfarewell: Пока\n"),
		},
	}

	tr := NewTranslator(EN)
	if err := tr.LoadFromFS(fsys, "locales"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	if got := tr.Translate(EN, "greeting"); got != "Hello" {
		t.Errorf("EN greeting = %q, want Hello", got)
	}
	if got := tr.Translate(RU, "farewell"); got != "Пока" {
		t.Errorf("RU farewell = %q, want Пока", got)
	}
}

func TestLoadFromBytes(t *testing.T) {
	t.Parallel()
	tr := NewTranslator(EN)

	yamlData := []byte("msg: loaded from bytes\n")
	if err := tr.LoadFromBytes("ja", yamlData); err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	if got := tr.Translate("ja", "msg"); got != "loaded from bytes" {
		t.Errorf("got %q, want 'loaded from bytes'", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tr := newTestTranslator(t)
	SetDefault(tr)
	defer SetDefault(nil)

	var failCount atomic.Int32
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()

			var ctx context.Context
			if id%2 == 0 {
				ctx = WithLocale(context.Background(), RU)
			} else {
				ctx = WithLocale(context.Background(), EN)
			}

			msg := TCtx(ctx, "greeting")
			if msg == "greeting" || msg == "" {
				failCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if n := failCount.Load(); n > 0 {
		t.Errorf("translation failed in %d goroutines", n)
	}
}

func TestFallback_Translator(t *testing.T) {
	t.Parallel()
	tr := NewTranslator(EN)

	if got := tr.Fallback(); got != EN {
		t.Errorf("Fallback() = %q, want EN", got)
	}
}

func TestNewTranslator_EmptyFallback(t *testing.T) {
	t.Parallel()
	tr := NewTranslator("")

	if got := tr.Fallback(); got != EN {
		t.Errorf("Fallback() with empty = %q, want EN", got)
	}
}

func TestDefault_Nil(t *testing.T) {
	old := defaultTranslator.Load()
	defaultTranslator.Store(nil)
	defer defaultTranslator.Store(old)

	if got := Default(); got != nil {
		t.Error("Default() should be nil when not set")
	}
}
