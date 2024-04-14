package artisum

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

type ArtisumLogHandler struct {
	name string
	callbacks.SimpleHandler
}

func NewArtisumLogHandler(name string) *ArtisumLogHandler {
	return &ArtisumLogHandler{name: name}
}

func (h *ArtisumLogHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	fmt.Println("------------------------")
	fmt.Printf("%s用のLLMの回答結果は以下になります\n", h.name)
	for _, c := range res.Choices {
		if c.Content != "" {
			fmt.Println("Content:", c.Content)
		}
		if c.StopReason != "" {
			fmt.Println("StopReason:", c.StopReason)
		}
		if len(c.GenerationInfo) > 0 {
			fmt.Println("GenerationInfo:")
			for k, v := range c.GenerationInfo {
				fmt.Printf("%20s: %v\n", k, v)
			}
		}
	}
	fmt.Println("------------------------")
}

func (h *ArtisumLogHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	fmt.Println("------------------------")
	fmt.Printf("%s用のChainの開始。以下の内容を入力として使います\n", h.name)
	for key, value := range inputs {
		fmt.Printf("\t%s: %v\n", key, value)
	}
	fmt.Println("------------------------")
}

func (h *ArtisumLogHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	fmt.Println("------------------------")
	fmt.Printf("%s用のChainの完了\n", h.name)
	fmt.Println("------------------------")
}
