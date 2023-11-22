package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// Integration test for SanityTranslateDocument,
// testing the whole flow from the request to the response.
// This test assumes that translation.metadata is empty.
func TestSanityTranslateDocumentTranslationMetadataEmpty(t *testing.T) {

	original_doc_id := "this-is-a-test-id"
	original_slug := "/en/this-is-a-test-slug"
	translated_slug := "/it/this-is-a-test-slug"

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/sanity_translate_document", SanityTranslateDocument)

	// Setup
	deleteTestDocument(original_doc_id + "_base")
	deleteTestDocument(original_doc_id + "_it")
	deleteTestDocument(original_doc_id)

	_, err := createTestDocument(original_doc_id, original_slug)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	// Payload for the translation request
	translationRequest := fmt.Sprintf(`
		{
			"FromLang": "en",
			"FromSlug": "%s",
			"ToLang": "it",
			"ToSlug": "%s",
			"InputElements": [
				"title"
			]
		}`, original_slug, translated_slug)

	req := httptest.NewRequest("POST", "/sanity_translate_document", bytes.NewBuffer([]byte(translationRequest)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check if the HTTP response is as expected
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, w.Code)
	}

	// Check if the title is as expected
	translated_query := fmt.Sprintf(`
		*[slug.current == '%s']{title}`,
		translated_slug,
	)
	r_translated_document, err := RunQuery(translated_query)
	title := gjson.Get(r_translated_document, "result.0.title").String()
	if title != "Questo è un titolo di prova" {
		t.Fatalf("Expected title to be 'Questo è un titolo di prova', got %v", title)
	}

	// Check if the translation.metadata is as expected
	metadata_query := fmt.Sprintf(`
		*[_id == '%s']`,
		original_doc_id+"_base",
	)
	r_metadata, err := RunQuery(metadata_query)

	// // Check for the Italian translation
	id_IT := gjson.Get(r_metadata, "result.0.translations.1.value._ref").String()
	IT_query := fmt.Sprintf(`
		*[_id == '%s']{"slug": slug.current}`,
		id_IT,
	)
	r_IT_document, err := RunQuery(IT_query)
	slug_IT := gjson.Get(r_IT_document, "result.0.slug").String()
	if slug_IT != translated_slug {
		t.Fatalf("Expected slug to be '%s', got %v", translated_slug, slug_IT)
	}

	// // Check for the English translation
	id_EN := gjson.Get(r_metadata, "result.0.translations.0.value._ref").String()
	EN_query := fmt.Sprintf(`
		*[_id == '%s']{"slug": slug.current}`,
		id_EN,
	)
	r_EN_document, err := RunQuery(EN_query)
	slug_EN := gjson.Get(r_EN_document, "result.0.slug").String()
	if slug_EN != original_slug {
		t.Fatalf("Expected slug to be '%s', got %v", original_slug, slug_EN)
	}

}

// Integration test for SanityTranslateDocument,
// testing the whole flow from the request to the response.
// This test assumes that translation.metadata is not empty (there has been a previous translation).
func TestSanityTranslateDocumentTranslationMetadataNotEmpty(t *testing.T) {

	original_doc_id := "this-is-a-test-id"
	original_slug := "/en/this-is-a-test-slug"
	translated_slug := "/it/this-is-a-test-slug"
	second_translated_slug := "/fr/this-is-a-test-slug"

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/sanity_translate_document", SanityTranslateDocument)

	// Setup
	deleteTestDocument(original_doc_id + "_base")
	deleteTestDocument(original_doc_id + "_it")
	deleteTestDocument(original_doc_id + "_fr")
	deleteTestDocument(original_doc_id)

	_, err := createTestDocument(original_doc_id, original_slug)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	_ = translated_slug

	// Payload for the first translation request
	firstTranslationRequest := fmt.Sprintf(`
		{
			"FromLang": "en",
			"FromSlug": "%s",
			"ToLang": "it",
			"ToSlug": "%s",
			"InputElements": [
				"title"
			]
		}`, original_slug, translated_slug)

	firstReq := httptest.NewRequest("POST", "/sanity_translate_document", bytes.NewBuffer([]byte(firstTranslationRequest)))
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, firstReq)

	// Payload for the second translation request
	secondTranslationRequest := fmt.Sprintf(`
		{
			"FromLang": "en",
			"FromSlug": "%s",
			"ToLang": "fr",
			"ToSlug": "%s",
			"InputElements": [
				"title"
			]
		}`, original_slug, second_translated_slug)

	secondReq := httptest.NewRequest("POST", "/sanity_translate_document", bytes.NewBuffer([]byte(secondTranslationRequest)))
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, secondReq)

	// Fetch the document translated
	second_translated_query := fmt.Sprintf(`
		*[slug.current == '%s']{title}`,
		second_translated_slug,
	)

	response, err := RunQuery(second_translated_query)
	title := gjson.Get(response, "result.0.title").String()

	// Check if the HTTP response is as expected
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, w1.Code)
	}

	// Check if the HTTP response is as expected
	if w2.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, w2.Code)
	}

	// Check if the title is as expected
	if title != "Il s'agit d'un titre de test" {
		t.Fatalf("Expected title to be 'Il s'agit d'un titre de test', got %v", title)
	}

	// Check if the translation.metadata is as expected
	metadata_query := fmt.Sprintf(`
		*[_id == '%s']`,
		original_doc_id+"_base",
	)
	r_metadata, err := RunQuery(metadata_query)

	// // Check for the English translation
	id_EN := gjson.Get(r_metadata, "result.0.translations.0.value._ref").String()
	EN_query := fmt.Sprintf(`
		*[_id == '%s']{"slug": slug.current}`,
		id_EN,
	)
	r_EN_document, err := RunQuery(EN_query)
	slug_EN := gjson.Get(r_EN_document, "result.0.slug").String()
	if slug_EN != original_slug {
		fmt.Println("CIAO 2")
		t.Fatalf("Expected slug to be '%s', got %v", original_slug, slug_EN)
	}

	// // Check for the Italian translation
	id_IT := gjson.Get(r_metadata, "result.0.translations.1.value._ref").String()
	IT_query := fmt.Sprintf(`
		*[_id == '%s']{"slug": slug.current}`,
		id_IT,
	)
	r_IT_document, err := RunQuery(IT_query)
	slug_IT := gjson.Get(r_IT_document, "result.0.slug").String()
	if slug_IT != translated_slug {
		fmt.Println("CIAO 1")
		t.Fatalf("Expected slug to be '%s', got %v", translated_slug, slug_IT)
	}

	// // Check for the French translation
	id_FR := gjson.Get(r_metadata, "result.0.translations.2.value._ref").String()
	FR_query := fmt.Sprintf(`
		*[_id == '%s']{"slug": slug.current}`,
		id_FR,
	)
	r_FR_document, err := RunQuery(FR_query)
	slug_FR := gjson.Get(r_FR_document, "result.0.slug").String()
	if slug_FR != second_translated_slug {
		fmt.Println("CIAO 3")
		t.Fatalf("Expected slug to be '%s', got %v", second_translated_slug, slug_FR)
	}
}

func createTestDocument(doc_id string, slug string) (string, error) {
	rawMutation := fmt.Sprintf(`
	 	{
	 		"mutations": [
	 			{
	 				"createOrReplace": {
						"_id": "%s",
						"language": "en",
						"_type": "test",
						"slug": {
							"_type": "slug",
							"current": "%s"
						},
						"title": "This is a test title",
						"intro": "This is an example of text",
						"portableTest": [
							{
								"_key": "c6280f5ed117",
								"_type": "block",
								"children": [
									{
										"_key": "e458c432651a",
										"_type": "span",
										"marks": [],
										"text": "This is a test string"
									}
								],
								"markDefs": [],
								"style": "normal"
							}
						],
						"testArray": [
							"This is the first test string",
							"This is the second test string"
						]
					}
				}
			]
		}`, doc_id, slug)

	err := RunMutation(rawMutation)
	if err != nil {
		return "", err
	}
	return doc_id, nil
}

func deleteTestDocument(doc_id string) error {
	rawMutation := fmt.Sprintf(`
		{
			"mutations": [
				{
					"delete": {
						"id": "%s"
					}
				}
			]
		}`, doc_id)

	err := RunMutation(rawMutation)
	if err != nil {
		return err
	}
	return nil
}
