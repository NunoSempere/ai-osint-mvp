package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	jsonschema "github.com/sashabaranov/go-openai/jsonschema"
)

// https://openai.com/api/pricing/
var GPT3_5_turbo string = "gpt-3.5-turbo-0125"
var GPT4_o string = "gpt-4o-2024-05-13"
var GPT4_turbo string = "gpt-4-turbo"
var GPT4_o_mini string = "gpt-4o-mini"

type OpenAIRequest struct {
	prompt string
	model  string
	token  string
}

func fetchOpenAIAnswer(req OpenAIRequest) (string, error) {
	client := openai.NewClient(req.token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: req.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.prompt,
				},
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	result := resp.Choices[0].Message.Content
	return result, nil
}


func fetchOpenAIAnswerJSON(req OpenAIRequest, schema openai.ChatCompletionResponseFormatJSONSchema) (string, error) {
	client := openai.NewClient(req.token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: req.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &schema,
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	result := resp.Choices[0].Message.Content
	return result, nil
}

type SummaryBox struct {
	Summary string  `json:"summary"`
	Error   *string `json:"error"`
}

func Summarize(text string, token string) (string, error) {
	prompt := "Please provide a mostly concise summary of the following Twitter activity. Focus on:\n" +
		"1. Key themes and topics discussed\n" +
		"2. Notable patterns in posting behavior\n" +
		"3. Important mentions or interactions\n" +
		"4. Overall sentiment and tone\n" +
		"5. Any significant events or announcements\n\n" +
	  "If there are any highly significant events or developments, spend a paragraph on them, but not more.\n\n" + 
		"Provide the response as JSON with this format: {\"summary\": \"your detailed summary here\", \"error\": null}\n\n" +
		"Twitter Activity Data:\n" + text


	var summary_box SummaryBox
	schema, err := jsonschema.GenerateSchemaForType(summary_box)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	openai_schema  := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Summary",
		Schema: schema,
		Strict: true,
	}
	summary_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT4_o_mini, token: token}, openai_schema)
	if err != nil {
		return "", err
	}
	
	err = json.Unmarshal([]byte(summary_json), &summary_box)
	if err != nil {
		log.Printf("Error unmarshalling json: %v", err)
		log.Printf("String was: %v", summary_json)
		return "", err
	}
	summary := summary_box.Summary
	return summary, nil
}

func TranslateString(text string, token string) (string, error) {
	prompt := "Translate this text into English: " + text + "\n"
	translation, err := fetchOpenAIAnswer(OpenAIRequest{prompt: prompt, model: GPT4_turbo, token: token})
	if err != nil {
		return "", err
	}
	translation_trimmed := strings.TrimSpace(translation)
	return translation_trimmed, nil
}

func MergeArticles(text string, token string) (string, error) {
	prompt := "Consider the following list of articles and their summaries. Your task is to clean it up.\n\n" +
		"1. If there are many articles, add a tl;dr at the top with the events which would most likely end up with > 1M deaths. Make this a paragraph starting with <p><b>tl;dr:</b>..., not an h1 element\n" +
		"2. Some of the articles may be talking about the same event—if so, join them together in one subsection, merge their summaries and reasoning, and create a list of the links that point to the same event. Otherwise, repeat the content of each item.\n" +
		"3. If there are any empty h1 headers (h1 headers followed immediately by another h1 header, skip those).\n" +
		"4. If do some other type of cleanup, point it out at the end.\n\n" +
		"Don't acknowledge instructions, just answer with the html.\n\n" + text

	summary, err := fetchOpenAIAnswer(OpenAIRequest{prompt: prompt, model: GPT4_turbo, token: token})
	if err != nil {
		return "", err
	}
	return summary + "<details><summary>The above articles were merged by GPT4-turbo. But you can view the originals under this toggle</summary>" + text + "</details>", nil
}
