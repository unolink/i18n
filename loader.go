package i18n

import (
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFromFS loads translations from an [fs.FS] (including [embed.FS]).
// Expected structure: dir/ru.yaml, dir/en.yaml, etc.
// The locale is extracted from the filename (without extension).
func (t *Translator) LoadFromFS(fsys fs.FS, dir string) error {
	return fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(stdpath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		filename := stdpath.Base(path)
		locale := Locale(strings.TrimSuffix(filename, stdpath.Ext(filename)))

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var translations map[string]string
		if err := yaml.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		t.dictionary.SetBatch(locale, translations)
		return nil
	})
}

// LoadFromFile loads translations from a single YAML file for a specific locale.
func (t *Translator) LoadFromFile(locale Locale, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	var translations map[string]string
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("parse %s: %w", filePath, err)
	}

	t.dictionary.SetBatch(locale, translations)
	return nil
}

// LoadFromBytes loads translations from YAML bytes for a specific locale.
func (t *Translator) LoadFromBytes(locale Locale, data []byte) error {
	var translations map[string]string
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	t.dictionary.SetBatch(locale, translations)
	return nil
}
