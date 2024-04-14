package artisum

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/textsplitter"
)

type InterestArticle struct {
	Title string
	Tag   string
	URL   string
}

type InterestTag struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type PromptArticle struct {
	FeedURL string
	URL     string
	Title   string
	Content string
}

type PromptResult struct {
	ExtractedArticles []*InterestArticle `json:"ExtractedArticles"`
}

type Extracter struct {
	modelName              string
	gpt35Turbo             *openai.LLM
	tags                   []*InterestTag
	mapReduceDocumentChain chains.MapReduceDocuments
	toJsonPromptTemplate   prompts.PromptTemplate
}

func NewExtracter(modelName string, numOfSummary int, tags []*InterestTag) (*Extracter, error) {
	llm, err := openai.New(
		openai.WithModel(modelName),
		openai.WithCallback(NewArtisumLogHandler("興味対象から記事抽出")))
	if err != nil {
		return nil, err
	}

	gpt35Turbo, err := openai.New(openai.WithModel("gpt-3.5-turbo"))
	if err != nil {
		return nil, err
	}

	tasgByte, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	llmPromptTemplate := prompts.NewPromptTemplate(fmt.Sprintf(`## Introduction
		You are an excellent engineer with extensive experience in web application development.

		Below, there are sections for "Areas of Technical Interest" and "Technical Articles".
		In "Areas of Technical Interest", there is a JSON array of objects, each containing fields for Name and Level.
		----json
		[
			{
				"Name": "Sample",
				"Level": 3
			}
		]
		----
		Name indicates the name of the interest, and Level indicates the degree of interest. The interest level is defined by numbers from 1 to 3, with higher numbers indicating greater interest.

		In "Technical Articles", there is a JSON array of objects, each containing fields for FeedURL, URL, Title, and Content.
		----json
		[
			{
				"FeedURL": "http://sample/feed/content.com",
				"URL": "https://sample.com",
				"Title": "sample",
				"Content": "samples content"
				"Description": "sample description",
			}
		]
		----

		You are to extract articles of high interest from the "Technical Articles" data based on the data from "Areas of Technical Interest".
		### Requirements for extraction:
		- The number of articles to extract should be under 2.
		- The number of articles to extract should be at least 1.
		- The articles selected must be highly relevant to the interests.
		- Choose articles that are important, versatile, and useful for web application product development.
		- The output format of the extracted data should be in the following JSON format:
		----json
		{
			"Title": "Sample",
			"Tag": "Sample",
			"URL": "https://sample.com",
		}
		----
		"Tag" should use the "Name" from "Areas of Technical Interest" directly.
		You have to get "Title", "ImageURL" value from "Technical Articles" directly.

		When outputting the response results, please use only the Jsonized data of the extracted articles as the output content.

		## "Technical Articles"
		{{.context}}

		## "Areas of Technical Interest"
		%s
		---
	`, tasgByte), []string{"context"})

	reducePromptTemplate := prompts.NewPromptTemplate(fmt.Sprintf(`
	You are to extract articles from "## Articles"
	### Requirements for extraction:
	- The number of articles to extract should be %d.
	- If there are %d or fewer articles, select all of them.
	- If there are more than %d, choose articles that are important, versatile, and useful for web application product development.
	- The output format of the extracted data should be in the following JSON format:
	{
		"ExtractedArticles": [
			{
				"Title": "Sample",
				"Tag": "Sample",
				"URL": "https://sample.com",
			}
			{
				"Title": "Sample2",
				"Tag": "Sample2",
				"URL": "https://sample2.com",
			}
		]
	}

	Ensure that "ExtractedArticles" value is always in array format, even if there is only one result.

	## Articles
	{{.context}}
	`, numOfSummary, numOfSummary, numOfSummary), []string{"context"})

	toJsonPromptTemplate := prompts.NewPromptTemplate(`
	Please convert the value of "##Json Data" into JSON format.
	Ensure that the JSON format is correct and not missing any data.
	## Json Data
	{{.context}}
	`, []string{"context"})

	llmChain := chains.NewLLMChain(llm, llmPromptTemplate, chains.WithCallback(NewArtisumLogHandler("独立して興味対象の記事抽出")))
	reduceChain := chains.NewLLMChain(llm, reducePromptTemplate, chains.WithCallback(NewArtisumLogHandler("記事抽出の最終まとめ")))
	mapReduceDocumentChain := chains.NewMapReduceDocuments(llmChain, reduceChain)

	return &Extracter{
		modelName:              modelName,
		gpt35Turbo:             gpt35Turbo,
		tags:                   tags,
		mapReduceDocumentChain: mapReduceDocumentChain,
		toJsonPromptTemplate:   toJsonPromptTemplate,
	}, nil
}

func (e *Extracter) Extract(ctx context.Context, articlesMap map[string][]*Article) ([]*InterestArticle, error) {
	var articles []*PromptArticle
	for feedURL, feedArticles := range articlesMap {
		articles = append(articles, lo.Map(feedArticles, func(a *Article, _ int) *PromptArticle {
			return &PromptArticle{
				FeedURL: feedURL,
				URL:     a.Url,
				Title:   a.Title,
				Content: a.Content,
			}
		})...)
	}
	promptArticles, err := json.Marshal(articles)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(promptArticles)
	docs, err := documentloaders.NewText(reader).LoadAndSplit(ctx,
		textsplitter.NewRecursiveCharacter(
			textsplitter.WithModelName(e.modelName),
			textsplitter.WithChunkSize(3000),
			textsplitter.WithChunkOverlap(300),
			textsplitter.WithSeparators([]string{"},", "],"}),
		),
	)
	if err != nil {
		return nil, err
	}

	result, err := e.mapReduceDocumentChain.Call(ctx, map[string]any{"input_documents": docs})
	if err != nil {
		return nil, err
	}

	resultPrompt, err := e.toJsonPromptTemplate.Format(map[string]any{"context": result["text"]})
	if err != nil {
		return nil, err
	}
	jsonResult, err := e.gpt35Turbo.Call(ctx, resultPrompt, llms.WithJSONMode())
	if err != nil {
		return nil, err
	}
	var promptResult *PromptResult
	err = json.Unmarshal([]byte(jsonResult), &promptResult)
	if err != nil {
		return nil, err
	}

	return promptResult.ExtractedArticles, nil
}
