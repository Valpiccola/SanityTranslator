package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

func RunDeepl(text string, from_lang string, to_lang string) (string, error) {
	params := url.Values{}
	params.Add("text", text)
	params.Add("source_lang", from_lang)
	params.Add("target_lang", to_lang)
	params.Add("preserve_formatting", "1")
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest(
		"POST",
		deeplAPIURL,
		body,
	)
	if err != nil {
		fmt.Println("Error creating Deepl request")
		return "", err
	}
	req.Header.Set(
		"Authorization",
		fmt.Sprintf(`DeepL-Auth-Key %s`, os.Getenv("DEEPL_TOKEN")),
	)
	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error executing Deepl request")
		fmt.Println("Status", resp.Status)
		fmt.Println("Body", string(bodyText))
		os.Exit(1)
	}

	translated_text := gjson.Get(
		string(bodyText),
		"translations.0.text",
	).Str

	if text[0:1] == " " {
		translated_text = " " + translated_text
	}
	if text[len(text)-1:] == " " {
		translated_text = translated_text + " "
	}
	return translated_text, nil
}
