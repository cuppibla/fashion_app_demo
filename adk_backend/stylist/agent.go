package stylist

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"goagents/fittingroom"
	"goagents/tools"

	"github.com/hashicorp/go-retryablehttp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
	"google.golang.org/adk/tool/loadartifactstool"
	"google.golang.org/genai"
)

//go:embed instructions.md
var instructions string

const stateKeyPreviousProducts = "previously_used_products"
const stateKeyUserImageStr = "user_base_image_str"

// InjectPreviousProducts is a BeforeModelCallback that reads previously selected
// product IDs from session state and injects a hint into the prompt so the LLM
// picks different products this time.
func InjectPreviousProducts(ctx agent.CallbackContext, req *model.LLMRequest) (*model.LLMResponse, error) {
	prev, err := ctx.State().Get(stateKeyPreviousProducts)
	if err != nil {
		// No previous state — first run, proceed normally.
		return nil, nil
	}
	prevStr, ok := prev.(string)
	if !ok || prevStr == "" {
		return nil, nil
	}

	slog.Info("Injecting previously used products into prompt", "products", prevStr)

	// Append a hint to the last user content so the LLM sees it.
	for i := len(req.Contents) - 1; i >= 0; i-- {
		if req.Contents[i].Role == "user" {
			req.Contents[i].Parts = append(req.Contents[i].Parts,
				genai.NewPartFromText(fmt.Sprintf(
					"IMPORTANT: You previously suggested these products: %s. You MUST pick DIFFERENT complementary products this time to give the user fresh, new outfit ideas. Do NOT reuse any of those product IDs.",
					prevStr)))
			break
		}
	}
	return nil, nil
}

// SaveSelectedProducts is an AfterModelCallback that parses the LLM's final
// JSON response, extracts the product IDs from all outfits, and saves them
// to session state so the next invocation can avoid them.
func SaveSelectedProducts(ctx agent.CallbackContext, resp *model.LLMResponse, respErr error) (*model.LLMResponse, error) {
	if resp == nil || resp.Content == nil {
		return nil, nil
	}
	for _, part := range resp.Content.Parts {
		if part.Text == "" {
			continue
		}
		ids := extractProductIDs(part.Text)
		if len(ids) > 0 {
			data, _ := json.Marshal(ids)
			slog.Info("Saving selected products to session state", "products", string(data))
			if err := ctx.State().Set(stateKeyPreviousProducts, string(data)); err != nil {
				slog.Warn("Failed to save previously used products to state", "err", err)
			}
		}
	}
	return nil, nil
}

// extractProductIDs parses a JSON string containing outfits and returns all
// product IDs found. It handles both raw JSON and markdown-fenced JSON.
func extractProductIDs(text string) []string {
	var parsed OutfitResponse
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		slog.Warn("Failed to parse product IDs from response", "err", err)
		return nil
	}

	var ids []string
	for _, outfit := range parsed.Outfits {
		for _, p := range outfit.Products {
			if p.ID != "" && !slices.Contains(ids, p.ID) {
				ids = append(ids, p.ID)
			}
		}
	}
	return ids
}

// NewStylistAgent creates an agent that acts as a fashion stylist,
// using the catalog agent as a tool to find items for the user.
func NewStylistAgent(apiKey string, catalogAgent agent.Agent) (agent.Agent, error) {
	c := retryablehttp.NewClient()
	ctx := context.Background()
	m, err := gemini.NewModel(ctx, "gemini-3-pro-preview", &genai.ClientConfig{
		APIKey:     apiKey,
		HTTPClient: c.StandardClient(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	fittingTool, err := fittingroom.NewFittingTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create fitting tool: %w", err)
	}

	imgtool, err := tools.NewImageTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create image tool: %w", err)
	}

	stylingAgent, err := llmagent.New(llmagent.Config{
		Name:        "stylist",
		Model:       m,
		Description: "A fashion stylist that suggests items based on user preferences",
		Instruction: instructions,
		SubAgents:   []agent.Agent{catalogAgent},
		Tools: []tool.Tool{
			agenttool.New(catalogAgent, nil),
			fittingTool,
			imgtool,
			loadartifactstool.New(),
		},
		BeforeAgentCallbacks: []agent.BeforeAgentCallback{
			fittingroom.SaveIncomingBlobs,
			tools.LogAgentInputCallback,
		},
		BeforeModelCallbacks: []llmagent.BeforeModelCallback{
			InjectPreviousProducts,
		},
		AfterModelCallbacks: []llmagent.AfterModelCallback{
			SaveSelectedProducts,
		},
		GenerateContentConfig: &genai.GenerateContentConfig{
			ResponseJsonSchema: adkStyleSchema(),
			ResponseMIMEType:   "application/json",
			ThinkingConfig: &genai.ThinkingConfig{
				IncludeThoughts: false,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create stylist agent: %w", err)
	}

	return stylingAgent, nil
}

// data types for response.
type Product struct {
	ID       string
	Title    string
	Subtitle string
	Price    float32
	image    string
}
type Outfit struct {
	Image      string
	Commentary string
	Products   []Product
}
type OutfitResponse struct {
	Outfits []Outfit `json:"outfit"`
}

// adkStyleSchema returns the messy map[string]any form of the json schema.
// This syntax is awkward, but it works.
func adkStyleSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"outfits": map[string]any{
				"type":     "array",
				"minItems": 3,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"image": map[string]any{
							"type":        "string",
							"description": "artifact name of a generated fitting image, as generated by the fitting tool.",
						},
						"commentary": map[string]any{
							"type": "string",
						},
						"products": map[string]any{
							"type":     "array",
							"minItems": 2,
							"maxItems": 5,
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"title": map[string]any{
										"type": "string",
									},
									"subtitle": map[string]any{
										"type": "string",
									},
									"price": map[string]any{
										"type": "number",
									},
									"id": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
		},
	}

}
