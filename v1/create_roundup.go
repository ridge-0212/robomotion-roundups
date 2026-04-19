package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/robomotionio/robomotion-go/message"
	"github.com/robomotionio/robomotion-go/runtime"
)

type CreateRoundup struct {
	runtime.Node `spec:"id=Robomotion.Roundups.Create,name=Create Roundup,icon=mdiFileDocumentEdit,color=#6C5CE7"`

	// === REQUIRED INPUTS ===
	InHeadline runtime.InVariable[string] `spec:"title=Headline,type=string,scope=Message,name=headline,messageScope,jsScope,customScope,description=The headline for the roundup article"`
	InTargetAudience runtime.InVariable[string] `spec:"title=Target Audience,type=string,scope=Message,name=targetAudience,messageScope,jsScope,customScope,description=Who is this roundup for"`
	InKeywords runtime.InVariable[string] `spec:"title=Keywords,type=string,scope=Message,name=keywords,messageScope,jsScope,customScope,description=Comma-separated keywords for the roundup"`

	// === OPTIONAL INPUTS ===
	OptProductsCount runtime.OptVariable[int] `spec:"title=Products Count,type=int,scope=Message,name=productsCount,value=5,messageScope,customScope,jsScope,description=Number of products to include (1-10)"`
	OptSearchQueries runtime.OptVariable[string] `spec:"title=Search Queries,type=string,scope=Message,name=searchQueries,messageScope,customScope,jsScope,description=Comma-separated product search queries"`
	OptAmazonASINs runtime.OptVariable[string] `spec:"title=Amazon ASINs,type=string,scope=Message,name=amazonAsins,messageScope,customScope,jsScope,description=Comma-separated Amazon product ASINs"`
	OptProductURLs runtime.OptVariable[string] `spec:"title=Product URLs,type=string,scope=Message,name=productUrls,messageScope,customScope,jsScope,description=Comma-separated product URLs"`

	// === OPTIONS ===
	OptProductType string `spec:"title=Product Type,value=unified,enum=amazon|appsumo|envato|unified,option,description=Product source platform"`
	OptTone string `spec:"title=Tone of Voice,value=professional,enum=professional|casual|friendly|formal|humorous|persuasive|academic|conversational|authoritative|enthusiastic|neutral,option,description=Writing style for the article"`
	OptLanguage string `spec:"title=Language,value=en,enum=ar|bg|cs|da|de|el|en|es|et|fi|fr|he|hi|hr|hu|id|it|ja|ko|lt|lv|ms|nl|no|pl|pt|ro|ru|sk|sl|sv|th|tl|tr|uk|vi|zh,option,description=Output language"`
	OptPointOfView string `spec:"title=Point of View,value=third_person,enum=first_person_singular|first_person_plural|third_person,option,description=Article perspective"`
	OptComparisonTable bool `spec:"title=Comparison Table,value=true,option,description=Include a comparison table in the article"`
	OptIncludePricing bool `spec:"title=Include Pricing,value=true,option,description=Include product pricing information"`
	OptIncludeRating bool `spec:"title=Include Rating,value=true,option,description=Include product ratings"`
	OptOptimizeFor string `spec:"title=Optimize Output For,value=seo,enum=seo|blog_post|social_media|email_newsletter|affiliate_marketing,option,description=Output optimization target"`
	OptLLMModel string `spec:"title=LLM Model,value=enhanced,enum=basic|enhanced,option,description=AI model quality level"`
	OptCoverImageStyle string `spec:"title=Cover Image Style,value=generative,enum=generative|product,option,description=Cover image generation style"`
	OptVisualStyle string `spec:"title=Visual Style,value=basic,enum=basic|charts,option,description=Visual presentation style"`
	OptLayoutStyle string `spec:"title=Layout Style,value=product_box,enum=product_box|showcase|youtube,option,description=Article layout style"`
	OptTemplateType string `spec:"title=Template Type,value=default,enum=default|awards,option,description=Article template type"`

	// === CREDENTIAL ===
	OptAPIKey runtime.Credential `spec:"title=API Key,scope=Custom,category=4,customScope,messageScope"`

	// === OUTPUTS ===
	OutRoundupID runtime.OutVariable[string] `spec:"title=Roundup ID,type=string,scope=Message,name=roundupId,messageScope"`
}

func (n *CreateRoundup) OnCreate() error { return nil }

func (n *CreateRoundup) OnMessage(ctx message.Context) error {
	headline, err := n.InHeadline.Get(ctx)
	if err != nil {
		return err
	}
	targetAudience, err := n.InTargetAudience.Get(ctx)
	if err != nil {
		return err
	}
	keywords, err := n.InKeywords.Get(ctx)
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

	productsCount, _ := n.OptProductsCount.Get(ctx)
	searchQueries, _ := n.OptSearchQueries.Get(ctx)
	amazonASINs, _ := n.OptAmazonASINs.Get(ctx)
	productURLs, _ := n.OptProductURLs.Get(ctx)

	body := map[string]interface{}{
		"headline":        headline,
		"target_audience": targetAudience,
		"keywords":        keywords,
	}

	if productsCount > 0 {
		body["products_count"] = productsCount
	}
	if searchQueries != "" {
		body["products_search_queries"] = searchQueries
	}
	if amazonASINs != "" {
		body["amazon_product_asins"] = amazonASINs
	}
	if productURLs != "" {
		body["product_urls"] = productURLs
	}
	if n.OptProductType != "" {
		body["product_type"] = n.OptProductType
	}
	if n.OptTone != "" {
		body["tone_of_voice"] = n.OptTone
	}
	if n.OptLanguage != "" {
		body["language"] = n.OptLanguage
	}
	if n.OptPointOfView != "" {
		body["point_of_view"] = n.OptPointOfView
	}
	body["comparison_table_enabled"] = n.OptComparisonTable
	body["include_pricing"] = n.OptIncludePricing
	body["include_rating"] = n.OptIncludeRating
	if n.OptOptimizeFor != "" {
		body["optimize_output_for"] = n.OptOptimizeFor
	}
	if n.OptLLMModel != "" {
		body["llm_model"] = n.OptLLMModel
	}
	if n.OptCoverImageStyle != "" {
		body["cover_image_style"] = n.OptCoverImageStyle
	}
	if n.OptVisualStyle != "" {
		body["visual_style"] = n.OptVisualStyle
	}
	if n.OptLayoutStyle != "" {
		body["layout_style"] = n.OptLayoutStyle
	}
	if n.OptTemplateType != "" {
		body["template_type"] = n.OptTemplateType
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return runtime.NewError("ErrInvalidJSON", "Failed to marshal request body: "+err.Error())
	}

	req, err := http.NewRequest("POST", "https://roundups.ai/api/v1/roundups", bytes.NewBuffer(jsonBody))
	if err != nil {
		return runtime.NewError("ErrRequestFailed", "Failed to create request: "+err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return runtime.NewError("ErrRequestFailed", "HTTP request failed: "+err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return runtime.NewError("ErrResponseRead", "Failed to read response: "+err.Error())
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return runtime.NewError("ErrRequestFailed", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return runtime.NewError("ErrInvalidJSON", "Failed to parse response JSON: "+err.Error())
	}

	roundupID, ok := result["id"].(string)
	if !ok {
		return runtime.NewError("ErrInvalidJSON", "Response missing roundup ID")
	}

	return n.OutRoundupID.Set(ctx, roundupID)
}

func (n *CreateRoundup) OnClose() error { return nil }
