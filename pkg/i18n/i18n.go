package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
)

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load all locale files
	files, _ := localeFS.ReadDir("locales")
	for _, f := range files {
		bundle.LoadMessageFileFS(localeFS, "locales/"+f.Name())
	}

	// Set default localizer
	SetLanguage("en")
}

// SetLanguage sets the current language for the localizer
func SetLanguage(lang string) {
	tag := language.Make(lang)
	localizer = i18n.NewLocalizer(bundle, tag.String())
}

// T translates a message by its ID and provides template data if needed
func T(id string, data interface{}) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil {
		return fmt.Sprintf("!%s!", id) // Return bracketed ID if not found
	}
	return msg
}
