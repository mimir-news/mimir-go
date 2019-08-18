package httpclient

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripQueryParameters(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input string
		ouput string
	}{
		{
			input: "http://interviewer:8080/v1/things?onlySome=1",
			ouput: "http://interviewer:8080/v1/things?onlySome=:value",
		},
		{
			input: "http://interviewer:8080/v1/things?onlySome=1&other=text-val",
			ouput: "http://interviewer:8080/v1/things?onlySome=:value&other=:value",
		},
		{
			input: "http://interviewer:8080/v1/things?onlySome=1&other=text-val&list=one,two,three",
			ouput: "http://interviewer:8080/v1/things?onlySome=:value&other=:value&list=:values",
		},
		{
			input: "http://interviewer:8080/v1/things",
			ouput: "http://interviewer:8080/v1/things",
		},
		{
			input: "",
			ouput: "",
		},
		{
			input: "http://interviewer:8080/v1/things?onlySome=1&other=text-val&list",
			ouput: "http://interviewer:8080/v1/things?onlySome=:value&other=:value&list=:value",
		},
		{
			input: "/v1/things?apiKey=secret-key&other=text-val&list=one,two,three",
			ouput: "/v1/things?apiKey=:value&other=:value&list=:values",
		},
		{
			input: "?apiKey=secret-key&other=text-val&list=one,two,three",
			ouput: "?apiKey=:value&other=:value&list=:values",
		},
	}

	for i, test := range tests {
		actual := stripQueryParameters(test.input)
		assert.Equal(test.ouput, actual, fmt.Sprintf("%d - stripQueryParameters failed", i+1))
	}
}

func TestStripQueryAndUUIDs(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input string
		ouput string
	}{
		{
			input: "http://interviewer:8080/v1/things/f0496797-0381-4b43-9821-313242a4d5f9?onlySome=1",
			ouput: "http://interviewer:8080/v1/things/:id",
		},
		{
			input: "http://interviewer:8080/v1/things/eff72a73-2a2c-41eb-9a8c-ea7f337c84d2/stuff",
			ouput: "http://interviewer:8080/v1/things/:id/stuff",
		},
		{
			input: "http://interviewer:8080/v1/things/397f50a8-a618-4c1b-8e9e-8b85e9412b6a/stuff/B84959F0-02B8-4AE9-BD2F-5EFE5C243D61",
			ouput: "http://interviewer:8080/v1/things/:id/stuff/:id",
		},
		{
			input: "http://interviewer:8080/v1/things/F88F9FCD-13AE-4BBB-B6A4-A539EC7FB804/stuff/9bb4c0af-f6da-4867-a72b-1a48b9bf014c/attr",
			ouput: "http://interviewer:8080/v1/things/:id/stuff/:id/attr",
		},
	}

	for i, test := range tests {
		actual := stripQueryAndUUIDs(test.input)
		assert.Equal(test.ouput, actual, fmt.Sprintf("%d - stripQueryAndUUIDs failed", i+1))
	}
}
