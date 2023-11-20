package main

import (
	sanity "github.com/sanity-io/client-go"
)

type SanityClient struct {
	Client *sanity.Client
}

// SanityTranslator holds the translation rules for Sanity documents.
type SanityTranslator struct {
	Id            string
	Lang          string        // Language to translate to
	FromSlug      string        // Slug of the document to translate
	ToSlug        string        // Slug of the translated document
	InputElements []string      // Elements to translate (e.g. text.000.children.000.text)
	Fields        []SanityField // Fields to translate (e.g. text.1.children.1.text)
	Before        string        // Document before any changes
	After         string        // Document after any changes
}

type SanityField struct {
	Path              string
	OriginalContent   string
	TranslatedContent string
}
