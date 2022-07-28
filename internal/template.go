package internal

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func FormatAsTitle(title string) string {
	title = cases.Title(language.Und, cases.NoLower).String(title)
	formatted := strings.ReplaceAll(title, "_", " ")
	return formatted
}

func FormatAsDate(createdAt time.Time) string {
	return createdAt.Format(time.RFC822)
}
