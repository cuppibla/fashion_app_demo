package fittingroom

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"goagents/tools"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"cloud.google.com/go/storage"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/adk/tool/loadartifactstool"
	"google.golang.org/genai"
)

//go:embed instructions.md
var instructions string

//go:embed tool_instructions.md
var toolInstructions string

type FittingToolArgs struct {
	UserImage   string   `json:"user_image" jsonschema:"the artifact name or gs:// URI of the user image (e.g. upload_xxx_1, a generated_fitting_... artifact name, or a gs:// URI from a prior fitting result)"`
	Accessories []string `json:"accessories" jsonschema:"list of artifact names of the product images to apply"`
}

type FittingToolResult struct {
	ArtifactName string `json:"artifact_name" jsonschema:"the artifact name of the generated image"`
	GCSUrl       string `json:"gcs_url,omitempty" jsonschema:"the gs:// URI of the generated image stored in GCS; use this as user_image for subsequent fitting calls"`
}

// gcsURIMimeType infers the MIME type from a GCS URI's file extension.
func gcsURIMimeType(uri string) string {
	switch strings.ToLower(filepath.Ext(uri)) {
	case ".png":
		return "image/png"
	default:
		return "image/jpeg"
	}
}

func doFitting(ctx tool.Context, args FittingToolArgs) (FittingToolResult, error) {
	log.Printf("FITTING TOOL: %+v", args)

	var userPart *genai.Part
	if strings.HasPrefix(args.UserImage, "gs://") {
		userPart = &genai.Part{FileData: &genai.FileData{
			FileURI:  args.UserImage,
			MIMEType: gcsURIMimeType(args.UserImage),
		}}
	} else {
		userImgResp, err := ctx.Artifacts().Load(ctx, args.UserImage)
		if err != nil {
			return FittingToolResult{}, err
		}
		userPart = userImgResp.Part
	}

	parts := []*genai.Part{
		genai.NewPartFromText(toolInstructions),
		genai.NewPartFromText("User Reference Photo:"),
		userPart,
	}

	var loaded int
	for _, acc := range args.Accessories {
		accResp, err := ctx.Artifacts().Load(ctx, acc)
		if err != nil {
			log.Printf("failed to load artifact %s : %v", acc, err)
			continue
		}
		loaded += 1
		parts = append(parts, genai.NewPartFromText("Accessory Images to Add:"))
		parts = append(parts, accResp.Part)
	}
	if loaded == 0 {
		return FittingToolResult{}, errors.New("no valid product images found")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
	})
	if err != nil {
		return FittingToolResult{}, err
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-image", []*genai.Content{genai.NewContentFromParts(parts, "user")}, &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE"},
		Temperature:        genai.Ptr(float32(0.15)),
	})
	if err != nil {
		return FittingToolResult{}, err
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return FittingToolResult{}, fmt.Errorf("no content generated")
	}

	// Find the image part — response may contain both TEXT and IMAGE parts
	var genPart *genai.Part
	for _, p := range resp.Candidates[0].Content.Parts {
		if p.InlineData != nil {
			genPart = p
			break
		}
	}
	if genPart == nil {
		return FittingToolResult{}, fmt.Errorf("no image generated")
	}

	artName := fmt.Sprintf("generated_fitting_%s_%s", ctx.InvocationID(), uuid.NewString()[:8])
	_, err = ctx.Artifacts().Save(ctx, artName, genPart)
	if err != nil {
		return FittingToolResult{}, err
	}

	// Also upload to GCS so the result can be used across sessions as a gs:// URI.
	gcsURI := ""
	if bucket := os.Getenv("GCS_BUCKET"); bucket != "" {
		ext := ".jpg"
		if genPart.InlineData.MIMEType == "image/png" {
			ext = ".png"
		}
		objectName := fmt.Sprintf("generated-fittings/%s%s", artName, ext)
		storageClient, serr := storage.NewClient(ctx)
		if serr != nil {
			slog.Warn("failed to create GCS client, skipping GCS upload", "err", serr)
		} else {
			defer storageClient.Close()
			w := storageClient.Bucket(bucket).Object(objectName).NewWriter(ctx)
			w.ContentType = genPart.InlineData.MIMEType
			if _, werr := w.Write(genPart.InlineData.Data); werr != nil {
				slog.Warn("failed to write fitting image to GCS", "err", werr)
				w.Close()
			} else if cerr := w.Close(); cerr != nil {
				slog.Warn("failed to close GCS writer", "err", cerr)
			} else {
				gcsURI = fmt.Sprintf("gs://%s/%s", bucket, objectName)
				slog.Info("saved fitting image to GCS", "uri", gcsURI)
			}
		}
	}

	return FittingToolResult{ArtifactName: artName, GCSUrl: gcsURI}, nil
}

type FittingRequest struct {
	UserPhoto     []byte       // artifact object name
	ProductPhotos []genai.Part // artifact or product filename
}

type FittingResponse struct {
	Img  genai.Image `jsonschema:"image data as base64 encoded bytes"`
	Text string      `jsonschema:"brief text description of the image"`
}

func NewFittingTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "fitting_tool",
		Description: "generates an image combining a user image and a list of accessory files, returns the artifact name of the generated image",
	}, doFitting)
}

func NewFittingRoomAgent(project string, catalogAgent agent.Agent) (agent.Agent, error) {
	ctx := context.Background()
	// Don't pass a custom HTTPClient — the genai SDK needs to wrap its own
	// OAuth transport around the client when using ADC. Passing one we made
	// ourselves bypasses that wrapping and produces 401 CREDENTIALS_MISSING.
	m, err := gemini.NewModel(ctx, "gemini-3.1-pro-preview", &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  project,
		Location: "global",
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	fittingTool, err := NewFittingTool()
	if err != nil {
		log.Fatalf("Failed to create fitting tool: %v", err)
	}

	imgtool, _ := tools.NewImageTool()

	fittingAgent, err := llmagent.New(llmagent.Config{
		Name:        "fitting room",
		Model:       m,
		Description: "Update an image to apply new items or accessories to a photo",
		Instruction: instructions,
		Tools: []tool.Tool{
			loadartifactstool.New(),
			imgtool,
			agenttool.New(catalogAgent, nil),
			fittingTool,
		},
		GenerateContentConfig: &genai.GenerateContentConfig{
			ResponseMIMEType:   "application/json",
			ResponseJsonSchema: fittingSchemaMap(),
			ThinkingConfig: &genai.ThinkingConfig{
				IncludeThoughts: false,
			},
		},
		BeforeAgentCallbacks: []agent.BeforeAgentCallback{
			SaveIncomingBlobs,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
		return nil, fmt.Errorf("failed to create agent %w", err)
	}
	return fittingAgent, nil
}

// SaveIncomingBlobs stores inline data from the user into Artifact storage
func SaveIncomingBlobs(ctx agent.CallbackContext) (*genai.Content, error) {
	slog.Info("saving incoming blobs", "r", ctx.UserContent())
	contents := ctx.UserContent()
	if len(contents.Parts) > 0 {
		for pindex, p := range contents.Parts {
			if p.InlineData != nil {
				aname := fmt.Sprintf("upload_%s_%d", ctx.InvocationID(), pindex)
				slog.Info("Saving inline data artifact", "filename", p.InlineData.DisplayName, "mimetype", p.InlineData.MIMEType, "artifact", aname)
				// save the incoming blob.
				if _, err := ctx.Artifacts().Save(ctx, aname, p); err != nil {
					slog.Error("Failed to save artifact", "error", err)
					return nil, err
				}
				return nil, nil
			}
		}
	}
	return nil, nil
}
func fittingSchemaMap() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"artifact_name": map[string]any{
				"type":        "string",
				"description": "artifact name of a generated fitting you created with fitting_tool",
			},
			"gcs_url": map[string]any{
				"type":        "string",
				"description": "URL to the image of a generated fitting",
			},
		},
	}

}
