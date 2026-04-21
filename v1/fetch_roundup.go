package v1

import (
	"encoding/json"

	"github.com/robomotionio/robomotion-go/message"
	"github.com/robomotionio/robomotion-go/runtime"
)

// FetchRoundup fetches the status and article content of a previously created roundup.
type FetchRoundup struct {
	runtime.Node `spec:"id=Robomotion.Roundups.FetchRoundup,name=Fetch Roundup,icon=mdiDownload,color=#6C5CE7"`

	// Credential
	OptAPIKey runtime.Credential `spec:"title=API Key,scope=Custom,category=4,messageScope,customScope"`

	// Input
	InRoundupID runtime.InVariable[int] `spec:"title=Roundup ID,type=int,scope=Message,name=roundupId,messageScope,jsScope,customScope,description=The ID of the roundup to fetch"`

	// Outputs
	OutState     runtime.OutVariable[string]      `spec:"title=State,type=string,scope=Message,name=state,messageScope"`
	OutHeadline  runtime.OutVariable[string]      `spec:"title=Headline,type=string,scope=Message,name=headline,messageScope"`
	OutTitle     runtime.OutVariable[string]      `spec:"title=Article Title,type=string,scope=Message,name=articleTitle,messageScope"`
	OutContent   runtime.OutVariable[string]      `spec:"title=Article Content,type=string,scope=Message,name=articleContent,messageScope"`
	OutImageURL  runtime.OutVariable[string]      `spec:"title=Featured Image URL,type=string,scope=Message,name=featuredImage,messageScope"`
	OutMetaDesc  runtime.OutVariable[string]      `spec:"title=Meta Description,type=string,scope=Message,name=metaDescription,messageScope"`
	OutResponse  runtime.OutVariable[interface{}] `spec:"title=Full Response,type=object,scope=Message,name=response,messageScope"`
	OutErrors    runtime.OutVariable[string]      `spec:"title=Errors,type=string,scope=Message,name=errors,messageScope"`
}

func (n *FetchRoundup) OnCreate() error { return nil }

func (n *FetchRoundup) OnMessage(ctx message.Context) error {
	// Get API key
	item, err := n.OptAPIKey.Get(ctx)
	if err != nil {
		return err
	}

	token, ok := item["value"].(string)
	if !ok || token == "" {
		return runtime.NewError("ErrInvalidArg", "Missing API Key value")
	}

	client := NewRoundupsClient(token)

	// Get roundup ID
	roundupID, err := n.InRoundupID.Get(ctx)
	if err != nil {
		return err
	}
	if roundupID <= 0 {
		return runtime.NewError("ErrInvalidArg", "Roundup ID must be a positive integer")
	}

	// Call API
	resp, err := client.FetchRoundup(roundupID)
	if err != nil {
		return err
	}

	// Set outputs
	if err := n.OutState.Set(ctx, resp.State); err != nil {
		return err
	}
	if err := n.OutHeadline.Set(ctx, resp.Headline); err != nil {
		return err
	}

	// Article fields (may be null if still generating)
	if resp.Article != nil {
		if err := n.OutTitle.Set(ctx, resp.Article.Title); err != nil {
			return err
		}
		if err := n.OutContent.Set(ctx, resp.Article.Content); err != nil {
			return err
		}
		if err := n.OutImageURL.Set(ctx, resp.Article.FeaturedImage); err != nil {
			return err
		}
		if err := n.OutMetaDesc.Set(ctx, resp.Article.MetaDescription); err != nil {
			return err
		}
	}

	// Errors field
	if resp.Errors != nil {
		if err := n.OutErrors.Set(ctx, *resp.Errors); err != nil {
			return err
		}
	}

	// Full response as JSON
	respJSON, _ := json.Marshal(map[string]interface{}{
		"id":         resp.ID,
		"headline":   resp.Headline,
		"state":      resp.State,
		"created_at": resp.CreatedAt,
		"updated_at": resp.UpdatedAt,
		"article":    resp.Article,
		"errors":     resp.Errors,
	})
	var respMap map[string]interface{}
	json.Unmarshal(respJSON, &respMap)
	if err := n.OutResponse.Set(ctx, respMap); err != nil {
		return err
	}

	return nil
}

func (n *FetchRoundup) OnClose() error { return nil }
