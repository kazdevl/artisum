package artisum

import (
	"context"
	"encoding/json"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type ArticleFormatter struct {
	formatLLM            *openai.LLM
	gpt35Turbo           *openai.LLM
	formatPromptTemplate prompts.PromptTemplate
	toJsonPromptTemplate prompts.PromptTemplate
}

type FormatContent struct {
	Heading   string   `json:"heading"`
	Sentences []string `json:"sentences"`
}

func NewArticleFormatter(modelName string) (*ArticleFormatter, error) {
	formatLLM, err := openai.New(
		openai.WithModel(modelName),
		openai.WithCallback(NewArtisumLogHandler("記事のフォーマット")),
	)
	if err != nil {
		return nil, err
	}
	gpt35Turbo, err := openai.New(openai.WithModel("gpt-3.5-turbo"))
	if err != nil {
		return nil, err
	}

	formatPromptBase :=
		`## Introduction
As an experienced IT engineer in web and game service development,
you will access the Content listed under "## Input Content".
Based on the content, please summarize and respond according to the items listed below.

1. Summary of the content
2. Key points in the content
3. Background and issues discussed in the article
4. Approaches taken regarding the background and issues
5. Results of the approaches
6. Technical keywords(please split the term with a comma if there are multiple terms)

And Please output the summary result in the following json format.
[
	{
		"heading": "Title", // please use above item
		"sentences": [ // please provide answers corresponding to the items listed above.
			"sentence1",
			"sentence2",
		]
	}
]

Ensure that results of 'senetences' are always returned in an array format, even if there is only one value or no value.
When converting to json format, please translate to Japanese, excluding the 6th item's "sentences".


## Input Content
{{.context}}`

	formatPromptTemplate := prompts.NewPromptTemplate(formatPromptBase, []string{"context"})

	toJsonPromptBase := `
	Please convert the value of "formatted_content" into JSON format.
	Ensure that the JSON format is correct and not missing any data.

	formatted_content
	{{.context}}
	`
	toJsonPromptTemplate := prompts.NewPromptTemplate(toJsonPromptBase, []string{"context"})

	return &ArticleFormatter{
		formatLLM:            formatLLM,
		gpt35Turbo:           gpt35Turbo,
		formatPromptTemplate: formatPromptTemplate,
		toJsonPromptTemplate: toJsonPromptTemplate,
	}, nil
}

func (f *ArticleFormatter) Format(ctx context.Context, url string) ([]*FormatContent, error) {
	textContent, err := ExtractTextContentFromURL(url)
	if err != nil {
		return nil, err
	}
	formatPrompt, err := f.formatPromptTemplate.Format(map[string]any{"context": textContent})
	if err != nil {
		return nil, err
	}
	resp, err := f.formatLLM.Call(ctx, formatPrompt)
	if err != nil {
		return nil, err
	}

	toJsonPrompt, err := f.toJsonPromptTemplate.Format(map[string]any{"context": resp})
	if err != nil {
		return nil, err
	}
	jsonResult, err := f.gpt35Turbo.Call(ctx, toJsonPrompt, llms.WithJSONMode())
	if err != nil {
		return nil, err
	}
	var formatArticle struct {
		Result []*FormatContent `json:"formatted_content"`
	}
	err = json.Unmarshal([]byte(jsonResult), &formatArticle)
	if err != nil {
		return nil, err
	}

	return formatArticle.Result, nil
}
