package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/robomotionio/robomotion-go/message"
	"github.com/robomotionio/robomotion-go/runtime"
)

type FetchRoundup struct {
	runtime.Node `spec:"id=Robomotion.Roundups.Fetch,name=Fetch Roundup,icon=mdiFileDownload,color=#6C5CE7"`

	// === INPUTS ===
	InRoundupID runtime.InVariable[string] `spec:"title=Roundup ID,type=string,scope=Message,name=roundupId,messageScope,jsScope,customScope,description=The roundup ID from Create Roundup"`

	// === OPTIONAL INPUTS ===
	OptTimeout runtime.OptVariable[int] `spec:"title=Poll Timeout (seconds),type=int,scope=Message,name=timeout,value=300,messageScope,customScope,jsScope,description=Max time to wait for generation (default 300s)"`
	OptPollInterval runtime.OptVariable[int] `spec:"title=Poll Interval (seconds),type=int,scope=Message,name=pollInterval,value=10,messageScope,customScope,jsScope,description=Seconds between status checks (default 10s)"`

	// === CREDENTIAL ===
	OptAPIKey runtime.Credential `spec:"title=API Key,scope=Custom,category=4,customScope,messageScope"`

	// === OUTPUTS ===
	OutTitle runtime.OutVariable[string] `spec:"title=Title,type=string,scope=Message,name=title,messageScope"`
	OutContent runtime.OutVariable[string] `spec:"title=Content,type=string,scope=Message,name=content,messageScope"`
	OutFeaturedImage runtime.OutVariable[string] `spec:"title=Featured Image,type=string,scope=Message,name=featuredImage,messageScope"`
	OutMetaDescription runtime.OutVariable[string] `spec:"title=Meta Description,type=string,scope=Message,name=metaDescription,messageScope"`
	OutState runtime.OutVariable[string] `spec:"title=State,type=string,scope=Message,name=state,messageScope"`
	OutArticle runtime.OutVariable[interface{}] `spec:"title=Article (Full),type=object,scope=Message,name=article,messageScope"`
}

func (n *FetchRoundup) OnCreate() error { return nil }

func (n *FetchRoundup) OnMessage(ctx message.Context) error {
	roundupID, err := n.InRoundupID.Get(ctx)
	if err != nil {
		return err
	}

	cred, err := n.OptAPIKey.Get(ctx)
	if err != nil {
		return err
	}
	if cred == nil {
		return runtime.NewError("ErrInvalidArg", "API Key is required")
	}
	apiKey, _ := cred["value"].(string)

	timeoutSec, _ := n.OptTimeout.Get(ctx)
	if timeoutSec <= 0 {
		timeoutSec = 300
	}
	pollSec, _ := n.OptPollInterval.Get(ctx)
	if pollSec <= 0 {
		pollSec = 10
	}

	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	url := fmt.Sprintf("https://roundups.ai/api/v1/roundups/%s", roundupID)
	client := &http.Client{}

	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return runtime.NewError("ErrRequestFailed", "Failed to create request: "+err.Error())
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return runtime.NewError("ErrRequestFailed", "HTTP request failed: "+err.Error())
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return runtime.NewError("ErrResponseRead", "Failed to read response: "+err.Error())
		}

		if resp.StatusCode != http.StatusOK {
			return runtime.NewError("ErrRequestFailed", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
		}

		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return runtime.NewError("ErrInvalidJSON", "Failed to parse response JSON: "+err.Error())
		}

		state, _ := result["state"].(string)

		if err := n.OutState.Set(ctx, state); err != nil {
			return err
		}

		if err := n.OutArticle.Set(ctx, result); err != nil {
			return err
		}

		if state == "draft" {
			article, _ := result["article"].(map[string]interface{})
			if article != nil {
				title, _ := article["title"].(string)
				content, _ := article["content"].(string)
				featuredImage, _ := article["featured_image"].(string)
				metaDesc, _ := article["meta_description"].(string)

				if err := n.OutTitle.Set(ctx, title); err != nil {
					return err
				}
				if err := n.OutContent.Set(ctx, content); err != nil {
					return err
				}
				if err := n.OutFeaturedImage.Set(ctx, featuredImage); err != nil {
					return err
				}
				if err := n.OutMetaDescription.Set(ctx, metaDesc); err != nil {
					return err
				}
			}
			return nil
		}

		if state == "timeout" {
			return runtime.NewError("ErrRequestFailed", "Roundup generation timed out on the server")
		}

		time.Sleep(time.Duration(pollSec) * time.Second)
	}

	return runtime.NewError("ErrRequestFailed", fmt.Sprintf("Polling timed out after %d seconds", timeoutSec))
}

func (n *FetchRoundup) OnClose() error { return nil }
