package i18n

import "sync"

// Dictionary stores translations for multiple locales.
// Thread-safe for concurrent reads and writes.
type Dictionary struct {
	store map[Locale]map[string]string
	mu    sync.RWMutex
}

// NewDictionary creates a new empty dictionary.
func NewDictionary() *Dictionary {
	return &Dictionary{
		store: make(map[Locale]map[string]string),
	}
}

// Set adds or updates a translation for a specific locale and key.
func (d *Dictionary) Set(locale Locale, key, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.store[locale] == nil {
		d.store[locale] = make(map[string]string)
	}
	d.store[locale][key] = value
}

// Get retrieves a translation for a specific locale and key.
// Returns empty string if not found.
func (d *Dictionary) Get(locale Locale, key string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if localeDict, ok := d.store[locale]; ok {
		return localeDict[key]
	}
	return ""
}

// Has reports whether a translation exists for a specific locale and key.
func (d *Dictionary) Has(locale Locale, key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if localeDict, ok := d.store[locale]; ok {
		_, exists := localeDict[key]
		return exists
	}
	return false
}

// SetBatch adds multiple translations for a locale at once.
func (d *Dictionary) SetBatch(locale Locale, translations map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.store[locale] == nil {
		d.store[locale] = make(map[string]string)
	}
	for key, value := range translations {
		d.store[locale][key] = value
	}
}

// Locale returns all translations for a specific locale.
// Returns a copy to prevent mutation.
func (d *Dictionary) Locale(locale Locale) map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]string)
	if localeDict, ok := d.store[locale]; ok {
		for k, v := range localeDict {
			result[k] = v
		}
	}
	return result
}

// Locales returns a list of all registered locales.
func (d *Dictionary) Locales() []Locale {
	d.mu.RLock()
	defer d.mu.RUnlock()

	locales := make([]Locale, 0, len(d.store))
	for locale := range d.store {
		locales = append(locales, locale)
	}
	return locales
}

// Clear removes all translations for a specific locale.
func (d *Dictionary) Clear(locale Locale) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.store, locale)
}

// ClearAll removes all translations for all locales.
func (d *Dictionary) ClearAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.store = make(map[Locale]map[string]string)
}
