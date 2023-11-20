package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// DEEPL_API_URL defines the URL for the Deepl API used for translation.
const deeplAPIURL = "https://api-free.deepl.com/v2/translate"

// SanityTranslateDocument handles the main logic for translating Sanity documents.
func SanityTranslateDocument(c *gin.Context) {

	var txx SanityTranslator

	err := errors.New("")

	// Create a Translator object adding all the info from the request
	if err = c.BindJSON(&txx); err != nil {
		c.String(http.StatusBadRequest, "Failed binding event to JSON")
		fmt.Println("Failed binding event to JSON")
		return
	}

	// Create a SanityDocument object adding all the info from Sanity API
	query := fmt.Sprintf(`*[slug.current == '%s'][0]`, txx.FromSlug)
	txx.Before, err = RunQuery(query)
	if err != nil || txx.Before == "" {
		c.String(http.StatusBadRequest, "Error extracting original_doc from Sanity")
		fmt.Println("Error extracting original_doc from Sanity")
		return
	}
	result := gjson.Get(txx.Before, "result").Raw
	txx.Before = result
	txx.After = result

	// Update current response with new info necessary to Sanity
	err = EvolveSanityResponse(&txx)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed evolving Sanity response")
		fmt.Println("Failed evolving Sanity response")
		return
	}

	// Execute all the translations required
	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(txx.Before), &m)
	err = ExecuteTranslation(&txx, m, "")
	if err != nil {
		c.String(http.StatusBadRequest, "Failed executing translations")
		fmt.Println("Failed executing translations")
		return
	}
	for _, element := range txx.Fields {
		txx.After, _ = sjson.Set(
			txx.After,
			element.Path,
			element.TranslatedContent,
		)
	}

	// Push document to Sanity
	newDocumentMutation := fmt.Sprintf(`
		{
			"mutations": [
				{
					"create": %s
				}
			]
		}`,
		txx.After,
	)
	err = RunMutation(newDocumentMutation)
	if err != nil {
		c.String(http.StatusBadRequest, "Pushing new document to Sanity")
		fmt.Println("Pushing new document to Sanity")
		return
	}

	// Update translation metadata
	err = ManageTranslationMetadata(&txx)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed managing translation metadata")
		fmt.Println("Failed managing translation metadata")
		return
	}
}

// EvolveSanityResponse updates the response with new info necessary to Sanity
func EvolveSanityResponse(txx *SanityTranslator) (err error) {

	// Set id
	old_id := gjson.Get(txx.Before, "_id").Str
	new_id := old_id + fmt.Sprintf(`_%s`, txx.Lang)
	txx.After, err = sjson.Set(
		txx.After,
		"_id",
		new_id,
	)

	// Set slug
	txx.After, err = sjson.Set(
		txx.After,
		"slug.current",
		txx.ToSlug,
	)

	// Set lang
	txx.After, err = sjson.Set(
		txx.After,
		"language",
		txx.Lang,
	)

	return err
}

// ExecuteTranslation executes the translation for the given path using Deepl API
func ExecuteTranslation(txx *SanityTranslator, val interface{}, path string) error {
	switch v := val.(type) {
	case map[string]interface{}:
		for key, subVal := range v {
			subPath := ""
			if path == "" {
				subPath = key
			} else {
				subPath = path + "." + key
			}
			err := ExecuteTranslation(txx, subVal, subPath)
			if err != nil {
				fmt.Println("Error while parsing fields")
				return err
			}
		}
	case []interface{}:
		for i, subVal := range v {
			subPath := path + "." + strconv.Itoa(i)
			err := ExecuteTranslation(txx, subVal, subPath)
			if err != nil {
				fmt.Println("Error while parsing fields")
				return err
			}
		}
	default:
		for _, translation := range txx.InputElements {
			if cleanString(translation) == cleanString(path) {
				if fmt.Sprintf("%v", v) == "" {
					continue
				}
				trax, err := RunDeepl(fmt.Sprintf("%v", v), txx.Lang)
				if err != nil {
					fmt.Println("Error while translating path")
					return err
				}
				txx.Fields = append(
					txx.Fields,
					SanityField{
						Path:              path,
						OriginalContent:   fmt.Sprintf("%v", v),
						TranslatedContent: trax,
					},
				)
			}
		}
	}
	return nil
}

// ManageTranslationMetadata updates the translation metadata document to keep reference in sync
func ManageTranslationMetadata(txx *SanityTranslator) (err error) {

	query := fmt.Sprintf(`*[slug.current == '%s']{
		"translation": *[
			_type == "translation.metadata" &&
			references(^._id)
		]
	}`, txx.FromSlug)

	document, err := RunQuery(query)
	if err != nil {
		fmt.Println("Error extracting translation.metadata from Sanity")
		return err
	}

	// If the document is empty it means no tranlsation.metadata document exists
	if gjson.Get(document, "result.0.translation").String() == "[]" {
		fmt.Println("Managing an empty translation.metadata document")

		rawMutation := fmt.Sprintf(`
	 	{
	 		"mutations": [
	 			{
	 				"create": {
	 					"_type": "translation.metadata",
	 					"translations": [
	 						{
	 							"_key": "it",
	 							"value": {
	 								"_ref": "%s",
	 								"_type": "reference"
	 							}
	 						},
	 						{
	 							"_key": "%s",
	 							"value": {
	 								"_ref": "%s",
	 								"_type": "reference"
	 							}
	 						}
	 					]
	 				}
	 			}
	 		]
	 	}`,
			gjson.Get(txx.Before, "_id").String(),
			txx.Lang,
			gjson.Get(txx.After, "_id").String(),
		)

		err = RunMutation(rawMutation)
		if err != nil {
			fmt.Println("Error running mutation")
			return err
		}
	} else { // If the document is not empty it means a tranlsation.metadata document exists

		languages := gjson.Get(document, "result.#.translation.#.translations.#._key")

		need_to_create := true
		for _, firstLevel := range languages.Array() {
			for _, secondLevel := range firstLevel.Array() {
				for _, lang := range secondLevel.Array() {
					if lang.Str == txx.Lang {
						need_to_create = false
					}
				}
			}
		}

		if need_to_create {

			fmt.Println("Managing a non-empty translation.metadata document")

			rawPatch := fmt.Sprintf(`
		 		{
		 			"mutations": [
		 				{
		 					"patch": {
		 						"id": "%s",
		 						"insert": {
		 							"after": "translations[-1]",
		 							"items": [
		 								{
											"_key": "%s",
											"value": {
												"_ref": "%s",
												"_type": "reference"
											}
										}
		 							]
		 						}
		 					}
		 				}
		 			]
		 		}`,
				gjson.Get(document, "result.0.translation.0._id").String(),
				txx.Lang,
				gjson.Get(txx.After, "_id").String(),
			)

			RunMutation(rawPatch)
			if err != nil {
				fmt.Println("ERROR PATCHING: %w", err)
			}

		}
	}

	return err

}

// cleanString removes all non-alphabetic characters from the input string.
func cleanString(text string) string {
	return regexp.MustCompile(`[^a-zA-Z]+`).ReplaceAllString(text, "")
}
