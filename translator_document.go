package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// SanityTranslateDocument handles the main logic for translating Sanity documents.
func SanityTranslateDocument(c *gin.Context) {

	var txx SanityDocumentTranslator

	err := errors.New("")

	// Create a Translator object adding all the info from the request
	if err = c.BindJSON(&txx); err != nil {
		c.String(http.StatusBadRequest, "Failed binding event to JSON")
		fmt.Println("Failed binding event to JSON")
		return
	}

	fmt.Printf("Translating from: %s\n", txx.FromSlug)

	// Create a SanityDocument object adding all the info from Sanity API
	query := fmt.Sprintf(`*[slug.current == '%s'][0]`, txx.FromSlug)
	txx.Before, err = RunQuery(query)
	if err != nil || txx.Before == "" {
		c.String(http.StatusBadRequest, "Error extracting original_doc from Sanity")
		fmt.Println("Error extracting original_doc from Sanity")
		return
	}
	result := gjson.Get(txx.Before, "result").Raw
	txx.Id = gjson.Get(result, "_id").String()
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
					"createOrReplace": %s
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

	fmt.Printf("Translating to: %s\n\n", txx.ToSlug)

	c.JSON(http.StatusOK, gin.H{
		"message": "Document translated successfully",
	})
}

// EvolveSanityResponse updates the response with new info necessary to Sanity
func EvolveSanityResponse(txx *SanityDocumentTranslator) (err error) {

	// Set id
	old_id := gjson.Get(txx.Before, "_id").Str
	new_id := old_id + fmt.Sprintf(`_%s`, txx.ToLang)
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
		txx.ToLang,
	)

	return err
}

// ExecuteTranslation executes the translation for the given path using Deepl API
func ExecuteTranslation(txx *SanityDocumentTranslator, val interface{}, path string) error {
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
				trax, err := RunDeepl(
					fmt.Sprintf("%v", v),
					txx.FromLang,
					txx.ToLang,
				)
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
func ManageTranslationMetadata(txx *SanityDocumentTranslator) error {
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

	languages := gjson.Get(document, "result.#.translation.#.translations.#._key")

	isEmpty := true
	for _, innerSlice := range languages.Array() {
		if len(innerSlice.Array()) != 0 {
			isEmpty = false
			break
		}
	}

	if isEmpty {

		// fmt.Println("Creating new translation.metadata document")

		rawMutation := fmt.Sprintf(`
			{
				"mutations": [
					{
						"create": {
							"_type": "translation.metadata",
							"_id": "%s",
							"translations": [
								{
									"_key": "%s",
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
			txx.Id+"_base",
			txx.FromLang,
			gjson.Get(txx.Before, "_id").String(),
			txx.ToLang,
			gjson.Get(txx.After, "_id").String(),
		)
		err = RunMutation(rawMutation)
		if err != nil {
			fmt.Printf("Error running mutation: %v\n", err)
			return err
		} else {
			return nil
		}

	} else {

		// fmt.Println("Updating existing translation.metadata document")

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
			txx.Id+"_base",
			txx.ToLang,
			gjson.Get(txx.After, "_id").String(),
		)
		err = RunMutation(rawPatch)
		if err != nil {
			fmt.Printf("Error running mutation: %v\n", err)
			return err
		}
	}

	return nil
}
