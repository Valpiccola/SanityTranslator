package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// convertSanityPathToGJSONPath converts Sanity paths to gjson compatible paths
func convertSanityPathToGJSONPath(sanityPath string) string {
	// Convert array accessors from [index] to .index
	re := regexp.MustCompile(`\[(\d+)\]`)
	return re.ReplaceAllString(sanityPath, ".$1")
}

// SanityTranslateField handles the main logic for translating a specific field in Sanity documents.
func SanityTranslateField(c *gin.Context) {
	var txx SanityFieldTranslator

	fmt.Println("Translating field")

	// Create a Translator object adding all the info from the request
	if err := c.BindJSON(&txx); err != nil {
		c.String(http.StatusBadRequest, "Failed binding event to JSON")
		fmt.Println("Failed binding event to JSON")
		return
	}

	// Create a SanityDocument object adding all the info from Sanity API
	query := fmt.Sprintf(`*[slug.current == '%s'][0]`, txx.FromSlug)
	originalDocument, err := RunQuery(query)
	if err != nil || originalDocument == "" {
		c.String(http.StatusBadRequest, "Error extracting original_doc from Sanity")
		fmt.Println("Error extracting original_doc from Sanity")
		return
	}
	result := gjson.Get(originalDocument, "result").Raw
	txx.Id = gjson.Get(result, "_id").String()
	txx.Before = result

	for _, mappingField := range txx.MappingFields {
		gjsonPath := convertSanityPathToGJSONPath(mappingField.SanityPath)

		fieldValue := gjson.Get(txx.Before, gjsonPath).String()
		if fieldValue == "" {
			c.String(http.StatusBadRequest, "Field not found in the document")
			fmt.Println("Field not found in the document")
			return
		}

		for _, toSlug := range txx.ToSlugs {
			query = fmt.Sprintf(`*[slug.current == '%s'][0]`, toSlug)
			translatedDoc, err := RunQuery(query)
			if err != nil {
				c.String(http.StatusBadRequest, "Error extracting translated_doc from Sanity")
				fmt.Println("Error extracting translated_doc from Sanity")
				return
			}
			translatedDocResult := gjson.Get(translatedDoc, "result").Raw
			translatedDocID := gjson.Get(translatedDocResult, "_id").String()

			translatedToLang := toSlug[1:3]
			translatedValue, err := RunDeepl(fieldValue, txx.FromLang, translatedToLang)
			if err != nil {
				c.String(http.StatusBadRequest, "Failed executing translation")
				fmt.Println("Failed executing translation")
				return
			}

			translatedValue = strings.TrimSpace(translatedValue)

			rawPatch := fmt.Sprintf(`
			{
				"mutations": [
					{
						"patch": {
							"id": "%s",
							"set": {
								"%s": "%s"
							}
						}
					}
				]
			}`,
				translatedDocID,
				mappingField.SanityPath,
				translatedValue,
			)

			err = RunMutation(rawPatch)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed patching translated field: %v", err)
				c.String(http.StatusBadRequest, errorMsg)
				fmt.Println(errorMsg)
				return
			}

			fmt.Printf("\tTranslating field: %s\n", toSlug)
		}
		fmt.Println("")
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"status":  "success",
			"message": "Field translation completed",
		},
	)
}
