package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type SanityResponse struct {
	Query  string `json:"query"`
	Result []struct {
		Title string `json:"title"`
	} `json:"result"`
}

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantTitle string
	}{
		{
			name:      "ValidQueryForTitle",
			query:     "*[slug.current == '/en/page/do-not-delete']{title}",
			wantTitle: "This is a test title",
		},
		{
			name:      "EmptyResponse",
			query:     "*[slug.current == '/en/not-existing-slug']{title}",
			wantTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			response, err := RunQuery(tt.query)

			var resp SanityResponse

			err = json.Unmarshal([]byte(response), &resp)
			if err != nil {
				t.Errorf("%s: RunQuery() = %v, want %v", tt.name, err, tt.wantTitle)
			}

			if tt.wantTitle == "" && len(resp.Result) == 0 {
				return
			}

			title := resp.Result[0].Title
			if title == tt.wantTitle {
				return
			}

			t.Errorf("%s: RunQuery() = %v, want %v", tt.name, len(resp.Result), 0)
		})
	}
}

func TestRunMutation(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	tests := []struct {
		name       string
		documentID string
	}{
		{
			name:       "ValidMutationForIntro",
			documentID: "1dc603f0-0012-4a1d-90e5-575b367804da",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			randomString := RandomString(10)
			expectedIntro := "Updated intro text " + randomString

			mutationData := fmt.Sprintf(`
		 		{
		 			"mutations": [
		 				{
		 					"patch": {
		 						"id": "%s",
		 						"set": {
		 							"intro": "%s"
		 						}
		 					}
		 				}
		 			]
		 		}`, tt.documentID, expectedIntro)

			err := RunMutation(mutationData)
			if err != nil {
				t.Fatalf("%s: RunMutation() returned an error: %v", tt.name, err)
			}

			response, err := RunQuery(fmt.Sprintf("*[_id == '%s']{intro}", tt.documentID))
			if err != nil {
				t.Fatalf("%s: RunQuery() returned an error when verifying mutation: %v", tt.name, err)
			}

			var resp struct {
				Result []struct {
					Intro string `json:"intro"`
				} `json:"result"`
			}
			err = json.Unmarshal([]byte(response), &resp)
			if err != nil {
				t.Fatalf("%s: Failed to unmarshal response: %v", tt.name, err)
			}

			if len(resp.Result) == 0 || resp.Result[0].Intro != expectedIntro {
				t.Errorf(
					"%s: Mutation did not result in the expected intro text, got: %v, want: %v",
					tt.name,
					resp.Result[0].Intro,
					expectedIntro,
				)
			}
		})
	}
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
