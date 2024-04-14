package artisum

import (
	"context"

	"github.com/jomei/notionapi"
)

type NotionRepository struct {
	client     *notionapi.Client
	databaseId notionapi.DatabaseID
}

func NewNotionRepository(token, databaseId string) *NotionRepository {
	return &NotionRepository{
		client:     notionapi.NewClient(notionapi.Token(token)),
		databaseId: notionapi.DatabaseID(databaseId),
	}
}

func (r *NotionRepository) SaveSummaryResult(ctx context.Context, article *SummaryArticle) error {
	_, err := r.client.Page.Create(ctx, r.createPageRequest(article))
	return err
}

func (r *NotionRepository) createPageRequest(article *SummaryArticle) *notionapi.PageCreateRequest {
	req := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: r.databaseId,
		},
		Properties: notionapi.Properties{
			"名前": notionapi.TitleProperty{
				Type: notionapi.PropertyTypeTitle,
				Title: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: article.Origin.Title,
						},
					},
				},
			},
			"タグ": notionapi.SelectProperty{
				Type: notionapi.PropertyTypeSelect,
				Select: notionapi.Option{
					Name: article.Origin.Tag,
				},
			},
			"記事": notionapi.URLProperty{
				Type: notionapi.PropertyTypeURL,
				URL:  article.Origin.URL,
			},
		},
		Children: r.createPageChildren(article.Contents),
	}

	return req
}

func (r *NotionRepository) createPageChildren(contents []*FormatContent) []notionapi.Block {
	var blocks []notionapi.Block
	for _, c := range contents {
		headingBlock := &notionapi.Heading2Block{
			Heading2: notionapi.Heading{
				Color: "green",
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: c.Heading,
						},
					},
				},
			},
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeHeading2,
			},
		}
		paragraphTexts := make([]notionapi.RichText, 0, len(c.Sentences))
		for _, s := range c.Sentences {
			paragraphTexts = append(paragraphTexts, notionapi.RichText{
				Type: notionapi.ObjectTypeText,
				Text: &notionapi.Text{
					Content: s,
				},
			})
		}
		paragraphBlock := &notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: paragraphTexts,
			},
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeParagraph,
			},
		}

		blocks = append(blocks, headingBlock, paragraphBlock)
	}
	return blocks
}
