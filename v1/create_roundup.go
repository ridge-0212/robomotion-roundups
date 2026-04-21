package v1

import (
	"encoding/json"

	"github.com/robomotionio/robomotion-go/message"
	"github.com/robomotionio/robomotion-go/runtime"
)

// CreateRoundup creates a new roundup and starts generation in the background.
type CreateRoundup struct {
	runtime.Node `spec:"id=Robomotion.Roundups.CreateRoundup,name=Create Roundup,icon=mdiNewspaperVariant,color=#6C5CE7"`

	// Credential
	OptAPIKey runtime.Credential `spec:"title=API Key,scope=Custom,category=4,messageScope,customScope"`

	// Input parameters (at least one of headline/targetAudience/keywords must be provided)
	OptHeadline          runtime.OptVariable[string]       `spec:"title=Headline,type=string,scope=Message,name=headline,messageScope,jsScope,customScope,description=Article headline for the roundup"`
	OptTargetAudience    runtime.OptVariable[string]       `spec:"title=Target Audience,type=string,scope=Message,name=targetAudience,messageScope,jsScope,customScope,description=Target audience description"`
	OptKeywords          runtime.OptVariable[string]       `spec:"title=Keywords,type=string,scope=Message,name=keywords,messageScope,jsScope,customScope,description=Keywords for product search"`

	// Product selection
	OptProductType        string `spec:"title=Product Type,value=,enum=|amazon|appsumo|envato|unified,enumNames=Auto|Amazon|AppSumo|Envato|Unified,option,description=Product source type. Auto detects from URLs"`
	OptProductsCount      int    `spec:"title=Products Count,value=0,option,description=Number of products to select. Zero means auto. Maximum 50"`
	OptProductsSearchQuery runtime.OptVariable[interface{}] `spec:"title=Search Queries,type=object,scope=Message,name=productsSearchQueries,messageScope,jsScope,customScope,description=Array of search queries for product selection. Maximum 10"`
	OptAmazonProductASINs  runtime.OptVariable[interface{}] `spec:"title=Amazon ASINs,type=object,scope=Message,name=amazonProductAsins,messageScope,jsScope,customScope,description=Array of Amazon ASINs. Maximum 50"`
	OptProductURLs         runtime.OptVariable[interface{}] `spec:"title=Product URLs,type=object,scope=Message,name=productUrls,messageScope,jsScope,customScope,description=Array of product URLs for unified type. Maximum 50"`

	// Style options
	OptToneOfVoice       string `spec:"title=Tone of Voice,value=,enum=|Conversational|Informative|Persuasive|Humorous|Inspirational|Reflective|Authoritative|Critical|Formal|Cautionary|Sarcastic,enumNames=Default|Conversational|Informative|Persuasive|Humorous|Inspirational|Reflective|Authoritative|Critical|Formal|Cautionary|Sarcastic,option,description=Writing tone for the article"`
	OptLanguage          string `spec:"title=Language,value=,enum=|English|English (UK)|German|Japanese|French|Italian|Spanish|Chinese (Simplified)|Chinese (Traditional)|Portuguese (Brazil)|Dutch|Korean|Russian|Arabic|Turkish|Polish|Swedish|Hindi,option,description=Output language"`
	OptPointOfView       string `spec:"title=Point of View,value=,enum=|second_person|first_person|third_person,enumNames=Default|Second Person|First Person|Third Person,option,description=Narrative perspective"`
	OptComparisonTable   bool   `spec:"title=Comparison Table,value=false,option,description=Include a comparison table"`
	OptIncludePricing    bool   `spec:"title=Include Pricing,value=false,option,description=Include product pricing information"`
	OptIncludeRating     bool   `spec:"title=Include Rating,value=false,option,description=Include product ratings"`
	OptLLMModel          string `spec:"title=LLM Model,value=basic,enum=basic|enhanced,enumNames=Basic|Enhanced,option,description=AI model quality"`
	OptCoverImageStyle   string `spec:"title=Cover Image Style,value=,enum=|generative|product,enumNames=Default|Generative|Product,option,description=OG cover image style"`
	OptVisualStyle       string `spec:"title=Visual Style,value=basic,enum=basic|charts,enumNames=Basic|Charts,option,description=Article visual style"`
	OptLayoutStyle       string `spec:"title=Layout Style,value=product_box,enum=product_box|showcase|youtube,enumNames=Product Box|Showcase|YouTube,option,description=Article layout style"`
	OptTemplateType      string `spec:"title=Template Type,value=default,enum=default|awards,enumNames=Default|Awards,option,description=Article template type"`
	OptOptimizeOutputFor string `spec:"title=Optimize For,value=,enum=|roundups_ai|wordpress|ghost_cms|email_distribution|medium,enumNames=Default|Roundups AI|WordPress|Ghost CMS|Email Distribution|Medium,option,description=Output optimization target"`
	OptCustomCTA         string `spec:"title=Custom CTA,value=,option,description=Custom call-to-action text"`
	OptProductCoverCount int    `spec:"title=Product Cover Count,value=0,option,description=Number of product images on OG cover. Range 1 to 10 when cover_image-style is product"`

	// Outputs
	OutRoundupID runtime.OutVariable[int]         `spec:"title=Roundup ID,type=int,scope=Message,name=roundupId,messageScope"`
	OutState     runtime.OutVariable[string]      `spec:"title=State,type=string,scope=Message,name=state,messageScope"`
	OutHeadline  runtime.OutVariable[string]      `spec:"title=Headline,type=string,scope=Message,name=headline,messageScope"`
	OutResponse  runtime.OutVariable[interface{}] `spec:"title=Full Response,type=object,scope=Message,name=response,messageScope"`
}

func (n *CreateRoundup) OnCreate() error { return nil }

func (n *CreateRoundup) OnMessage(ctx message.Context) error {
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

	// Build request
	req := &CreateRoundupRequest{}

	// Input parameters (at least one)
	if headline, err := n.OptHeadline.Get(ctx); err == nil && headline != "" {
		req.Headline = headline
	}
	if audience, err := n.OptTargetAudience.Get(ctx); err == nil && audience != "" {
		req.TargetAudience = audience
	}
	if keywords, err := n.OptKeywords.Get(ctx); err == nil && keywords != "" {
		req.Keywords = keywords
	}

	// Validate at least one required parameter
	if req.Headline == "" && req.TargetAudience == "" && req.Keywords == "" {
		return runtime.NewError("ErrInvalidArg", "At least one of Headline, Target Audience, or Keywords must be provided")
	}

	// Product selection
	if n.OptProductType != "" {
		req.ProductType = n.OptProductType
	}
	if n.OptProductsCount < 0 {
		return runtime.NewError("ErrInvalidArg", "Products Count must be non-negative")
	}
	if n.OptProductsCount > 0 {
		if n.OptProductsCount > 50 {
			return runtime.NewError("ErrInvalidArg", "Products Count must not exceed 50")
		}
		req.ProductsCount = n.OptProductsCount
	}

	// Parse array inputs — handles both []interface{} and []string
	if searchQueries, err := n.OptProductsSearchQuery.Get(ctx); err == nil && searchQueries != nil {
		queries, err := parseStringArray(searchQueries)
		if err != nil {
			return runtime.NewError("ErrInvalidArg", "Search Queries must be an array of strings")
		}
		req.ProductsSearchQuery = queries
		if len(req.ProductsSearchQuery) > 10 {
			return runtime.NewError("ErrInvalidArg", "Search Queries must not exceed 10 items")
		}
	}

	if asins, err := n.OptAmazonProductASINs.Get(ctx); err == nil && asins != nil {
		asinList, err := parseStringArray(asins)
		if err != nil {
			return runtime.NewError("ErrInvalidArg", "Amazon ASINs must be an array of strings")
		}
		req.AmazonProductASINs = asinList
		if len(req.AmazonProductASINs) > 50 {
			return runtime.NewError("ErrInvalidArg", "Amazon ASINs must not exceed 50 items")
		}
	}

	if urls, err := n.OptProductURLs.Get(ctx); err == nil && urls != nil {
		urlList, err := parseStringArray(urls)
		if err != nil {
			return runtime.NewError("ErrInvalidArg", "Product URLs must be an array of strings")
		}
		req.ProductURLs = urlList
		if len(req.ProductURLs) > 50 {
			return runtime.NewError("ErrInvalidArg", "Product URLs must not exceed 50 items")
		}
	}

	// Validation: unified type requires product_urls
	if req.ProductType == "unified" && len(req.ProductURLs) == 0 {
		return runtime.NewError("ErrInvalidArg", "Product URLs are required when Product Type is Unified")
	}

	// Style options
	styles := &StyleOptions{}
	hasStyles := false

	if n.OptToneOfVoice != "" {
		styles.ToneOfVoice = n.OptToneOfVoice
		hasStyles = true
	}
	if n.OptLanguage != "" {
		styles.Language = n.OptLanguage
		hasStyles = true
	}
	if n.OptPointOfView != "" {
		switch n.OptPointOfView {
		case "second_person":
			styles.PointOfView = "Second Person (You)"
		case "first_person":
			styles.PointOfView = "First Person (I, We)"
		case "third_person":
			styles.PointOfView = "Third Person (He, She, They, It)"
		default:
			styles.PointOfView = n.OptPointOfView
		}
		hasStyles = true
	}
	if n.OptComparisonTable {
		styles.ComparisonTable = boolPtr(true)
		hasStyles = true
	}
	if n.OptIncludePricing {
		styles.IncludePricing = boolPtr(true)
		hasStyles = true
	}
	if n.OptIncludeRating {
		styles.IncludeRating = boolPtr(true)
		hasStyles = true
	}
	if n.OptLLMModel != "" {
		styles.LLMModel = n.OptLLMModel
		hasStyles = true
	}
	if n.OptCoverImageStyle != "" {
		styles.CoverImageStyle = n.OptCoverImageStyle
		hasStyles = true
	}
	if n.OptVisualStyle != "" {
		styles.VisualStyle = n.OptVisualStyle
		hasStyles = true
	}
	if n.OptLayoutStyle != "" {
		styles.LayoutStyle = n.OptLayoutStyle
		hasStyles = true
	}
	if n.OptTemplateType != "" {
		styles.TemplateType = n.OptTemplateType
		hasStyles = true
	}
	if n.OptOptimizeOutputFor != "" {
		styles.OptimizeOutputFor = n.OptOptimizeOutputFor
		hasStyles = true
	}
	if n.OptCustomCTA != "" {
		styles.CustomCTA = n.OptCustomCTA
		hasStyles = true
	}
	if n.OptProductCoverCount < 0 {
		return runtime.NewError("ErrInvalidArg", "Product Cover Count must be non-negative")
	}
	if n.OptProductCoverCount > 0 {
		if n.OptCoverImageStyle != "product" {
			return runtime.NewError("ErrInvalidArg", "Product Cover Count only applies when Cover Image Style is set to product")
		}
		if n.OptProductCoverCount > 10 {
			return runtime.NewError("ErrInvalidArg", "Product Cover Count must be between 1 and 10")
		}
		styles.ProductCount = n.OptProductCoverCount
		hasStyles = true
	}

	if hasStyles {
		req.Styles = styles
	}

	// Call API
	resp, err := client.CreateRoundup(req)
	if err != nil {
		return err
	}

	// Set outputs
	if err := n.OutRoundupID.Set(ctx, resp.ID); err != nil {
		return err
	}
	if err := n.OutState.Set(ctx, resp.State); err != nil {
		return err
	}
	if err := n.OutHeadline.Set(ctx, resp.Headline); err != nil {
		return err
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

func (n *CreateRoundup) OnClose() error { return nil }

func boolPtr(b bool) *bool {
	return &b
}
