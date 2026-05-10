# 👗 Build a Virtual Fitting Room & AI Stylist with Flutter, ADK Go & Gemini


---


## Introduction
**Duration: 3 min**


### What You'll Build


In this codelab, you'll step into the shoes of a developer at **Thread Count**, a fictional retail brand with an existing Flutter shopping app. Your mission: add two AI-powered features that transform the online shopping experience.


1. **Virtual Fitting Room** — A user uploads a photo of themselves, selects a clothing item, and sees an AI-generated image of themselves wearing that item.
2. **AI Stylist** — Based on the user's location, occasion, and style preferences, an AI agent curates complete outfit recommendations — and the user can refine them through conversation.


The idea is simple: when people try clothes on in a fitting room, they're far more likely to buy them. But online? You're just guessing. This project bridges that gap with AI.


### Architecture at a Glance


```
Flutter App  ──── HTTP/REST ────▶  ADK Go Backend
                                       │
                            ┌──────────┼──────────┐
                       Fitting Room  Stylist    Catalog
                         Agent       Agent      Agent
                                       │
                            Gemini API + Cloud Storage
```


### Core Technologies


| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Agent Framework** | ADK (Agent Development Kit) for Go | Multi-agent orchestration, sessions, artifacts |
| **Agent Reasoning** | Gemini 3 Pro Preview | Powers the fitting room and stylist agents |
| **Image Generation** | Gemini 2.5 Flash Image | Generates try-on and outfit images |
| **Frontend** | Flutter (Dart) | Cross-platform app (Web, iOS, Android) |
| **Storage** | Google Cloud Storage | Stores product images and generated artifacts |
| **Hosting** | Cloud Run | Serverless container deployment |


---


## 📦 Prerequisites & Cloud Shell Setup
**Duration: 5 min**


### 1. Open Cloud Shell Editor


👉 Open [Cloud Shell Editor](https://ide.cloud.google.com/) in your browser.


If the terminal doesn't appear at the bottom of the screen:
- Click **View** → **Terminal**


> aside positive
> **Cloud Shell** comes with Go, Git, and the `gcloud` CLI pre-installed — no local setup needed.


### 2. Install Flutter SDK


Cloud Shell doesn't include Flutter by default. Install it in your home directory so it persists across sessions:


```bash
cd ~
git clone https://github.com/flutter/flutter.git -b stable --depth 1
export PATH="$PATH:$HOME/flutter/bin"
```


Make the PATH persistent:


```bash
echo 'export PATH="$PATH:$HOME/flutter/bin"' >> ~/.bashrc
source ~/.bashrc
```


Verify the installation:


```bash
flutter --version
```


### 3. Clone the Repository


```bash
cd ~
git clone https://github.com/anthropics/thread-count-workshop.git
cd thread-count-workshop
```


### 4. Explore the Project Structure


```
thread-count-workshop/
├── adk_backend/                 # Go backend with ADK agents
│   ├── main.go                  # Entry point — wires all agents + REST server
│   ├── catalog/                 # Catalog Agent — product lookup
│   │   ├── agent.go
│   │   ├── catalog.yaml         # Product database (YAML)
│   │   └── instructions.md      # Agent persona prompt
│   ├── fittingroom/             # Fitting Room Agent — virtual try-on
│   │   ├── agent.go
│   │   ├── instructions.md      # Agent persona prompt
│   │   └── tool_instructions.md # Image generation prompt
│   ├── stylist/                 # Stylist Agent — outfit curation
│   │   ├── agent.go
│   │   └── instructions.md      # Agent persona + output format
│   ├── rootagent/               # Root Agent — routes to the right agent
│   │   └── agent.go
│   └── tools/                   # Shared tools
│       ├── imagetool.go         # getProductImage — loads product images
│       ├── outfit_gen_tool.go   # generate_outfit_image — creates outfit images
│       └── cors_helper.go      # CORS middleware + request logging
│
├── flutter_frontend/            # Flutter cross-platform app
│   ├── lib/
│   │   ├── main.dart            # App entry point + Provider setup
│   │   ├── app_config.dart      # Backend URL configuration
│   │   ├── core_app/            # Pre-built shopping app (browse, cart, etc.)
│   │   └── workshop_tasks/      # AI feature code
│   │       ├── step_1_try_it_on/  # Virtual Try-On flow
│   │       │   ├── providers/     # TryItOnProvider (state management)
│   │       │   ├── services/      # AdkFittingRoomService (HTTP calls)
│   │       │   └── ui/            # Screens (product detail → try on → fitting room)
│   │       └── step_2_style_me/   # Style Me flow
│   │           ├── models/        # Outfit, StyleRequest data classes
│   │           ├── providers/     # StylingProvider (state management)
│   │           ├── services/      # AdkStylingService (HTTP calls)
│   │           └── ui/            # Screens (form sheet → outfit carousel)
│   └── assets/images/           # Product catalog images
```


> aside positive
> **Key directories:**
> - `adk_backend/` — Where the AI agents live. Each agent has its own package with Go code + markdown instructions.
> - `flutter_frontend/lib/workshop_tasks/` — Where the Flutter UI connects to those agents.
> - `flutter_frontend/lib/core_app/` — The pre-built shopping app. You won't need to modify this.


---


## ☁️ Google Cloud Project Setup
**Duration: 5 min**


### 1. Create a New Project


```bash
gcloud projects create thread-count-lab --name="Thread Count Lab"
gcloud config set project thread-count-lab
```


> aside positive
> If the project ID is taken, append a random number (e.g., `thread-count-lab-42`).


### 2. Enable Required APIs


```bash
gcloud services enable \
 aiplatform.googleapis.com \
 storage.googleapis.com \
 run.googleapis.com \
 cloudbuild.googleapis.com \
 artifactregistry.googleapis.com
```


| API | Purpose |
|-----|---------|
| `aiplatform.googleapis.com` | **Vertex AI** — the `fitting_tool` calls Gemini's image generation through Vertex AI |
| `storage.googleapis.com` | **Cloud Storage** — stores product catalog images and generated try-on results |
| `run.googleapis.com` | **Cloud Run** — hosts the backend as a serverless container |
| `cloudbuild.googleapis.com` | **Cloud Build** — builds Docker images from source |
| `artifactregistry.googleapis.com` | **Artifact Registry** — stores built Docker images |


### 3. Get a Gemini API Key


👉 Go to [Google AI Studio](https://aistudio.google.com/apikey) and create an API key.


> aside negative
> The backend uses **two different authentication methods**: the Gemini API key (for agent reasoning) and Vertex AI project auth (for image generation via `fitting_tool`). Both are needed.


### 4. Create a GCS Bucket


```bash
export PROJECT_ID=$(gcloud config get-value project)
gsutil mb gs://thread-count-$PROJECT_ID
```


### 5. Upload Product Catalog Images


```bash
cd ~/thread-count-workshop
gcloud storage cp -r flutter_frontend/assets/images/* \
 gs://thread-count-$PROJECT_ID/catalog-assets/images/
```


### 6. Configure the `.env` File


```bash
cd ~/thread-count-workshop/adk_backend
cat > .env << EOF
GEMINI_API_KEY=your_api_key_here
GOOGLE_CLOUD_PROJECT=$PROJECT_ID
GCS_BUCKET=thread-count-$PROJECT_ID
EOF
```


👉 Replace `your_api_key_here` with your actual Gemini API key.


> aside positive
> The `.env` file is gitignored — never commit API keys to source control.


### 7. Authenticate for Vertex AI


The `fitting_tool` uses Vertex AI (not the API key) for image generation. Authenticate `gcloud`:


```bash
gcloud auth application-default login
```


---


## 🏗️ Architecture Overview
**Duration: 5 min**


Now that the environment is ready, let's understand how the system works before looking at the code.


### The Four-Agent System


The backend is built as a **multi-agent system** using ADK (Agent Development Kit) for Go. Four agents work together, each with a specific responsibility:


```
                   ┌─────────────┐
                   │ Root Agent  │  ← Routes requests to the right agent
                   │ (gemini-2.5-│
                   │   flash)    │
                   └──────┬──────┘
                          │
           ┌──────────────┼──────────────┐
           ▼              ▼              ▼
  ┌────────────────┐ ┌──────────┐ ┌────────────────┐
  │ Fitting Room   │ │ Catalog  │ │ Stylist        │
  │ Agent          │ │ Agent    │ │ Agent          │
  │ (gemini-3-pro) │ │ (gemini- │ │ (gemini-3-pro) │
  │                │ │ 2.5-flash│ │                │
  │ Tools:         │ │          │ │ Tools:         │
  │ • fitting_tool │ │ Tools:   │ │ • fitting_tool │
  │ • getProduct   │ │ • list   │ │ • getProduct   │
  │   Image        │ │ Products │ │   Image        │
  │ • catalog_agent│ │ • get    │ │ • catalog_agent│
  │   (delegation) │ │ Product  │ │   (delegation) │
  │                │ │  Image   │ │ • generate_    │
  └────────────────┘ └──────────┘ │   outfit_image │
                                  └────────────────┘
```


| Agent | Model | Role |
|-------|-------|------|
| **Root Agent** | `gemini-2.5-flash` | Traffic cop. Reads the user's message and delegates to the right specialist agent. Uses a fast, lightweight model because it only needs to make routing decisions. |
| **Catalog Agent** | `gemini-2.5-flash` | Product expert. Loads the product catalog from a YAML file and answers product queries. Also lightweight — it's just looking up data. |
| **Fitting Room Agent** | `gemini-3-pro-preview` | Virtual try-on specialist. Takes a user photo + product image and generates a composite image of the person wearing that item. Uses a more capable model because it needs to reason about images. |
| **Stylist Agent** | `gemini-3-pro-preview` | Fashion advisor. Given location, occasion, and preferences, it curates 3 outfit combinations from the catalog. Can generate try-on images for each outfit. Also uses the capable model for creative reasoning. |


### The Entry Point: `main.go`


Everything starts in `main.go`, which wires the agents together and starts the HTTP server:


```go
// main.go — simplified for clarity


func main() {
   godotenv.Load()  // Load .env file


   // 1. Create the artifact storage (GCS-backed)
   artifacts, _ := gcsartifact.NewService(ctx, bucket)


   // 2. Build agents bottom-up (dependencies first)
   catagent, _    := catalog.NewCatalogAgent(apikey, "catalog/catalog.yaml")
   fitagent, _    := fittingroom.NewFittingRoomAgent(apikey, catagent)
   stylistAgent, _ := stylist.NewStylistAgent(apikey, catagent)
   ragent, _      := rootagent.NewRootAgent(apikey, fitagent, catagent, stylistAgent)


   // 3. Register all agents with a multi-loader
   loader, _ := agent.NewMultiLoader(ragent, fitagent, catagent, stylistAgent)


   // 4. Create the ADK REST server
   restHandler, _ := adkrest.NewServer(adkrest.ServerConfig{
       SessionService:  session.InMemoryService(),
       MemoryService:   memory.InMemoryService(),
       AgentLoader:     loader,
       ArtifactService: artifacts,
   })


   // 5. Mount behind /api/ with CORS support
   r := mux.NewRouter()
   r.Use(tools.LocalhostCORS)
   r.PathPrefix("/api/").Handler(
       http.StripPrefix("/api", tools.LogHandler(restHandler)))


   http.Server{Addr: ":8080", Handler: r}.ListenAndServe()
}
```


A few important things to notice:


- **Agents are built bottom-up**: The catalog agent is created first because both the fitting room and stylist agents depend on it (they delegate product lookups to it).
- **`agent.NewMultiLoader`** registers all four agents so the REST API can route to any of them by name.
- **`adkrest.NewServer`** provides the REST API automatically — you don't write endpoint handlers yourself. ADK gives you session management, artifact storage, and agent execution out of the box.
- **`session.InMemoryService()`** stores sessions in memory. This means sessions are lost if the server restarts, which is fine for a demo. In production, you'd use a persistent store.
- **`gcsartifact.NewService`** stores artifacts (generated images) in Google Cloud Storage, so they persist across requests and can be shared via GCS URIs.


---


## 🤖 ADK (Agent Development Kit) Deep Dive
**Duration: 8 min**


### What is ADK?


The **Agent Development Kit (ADK)** is an open-source framework from Google for building AI agents in Go (and Python/Java). It's the layer between your application and the Gemini API.


You *could* call the Gemini API directly. But once your app needs to:
- Look up products from a catalog
- Generate images based on user photos
- Remember what outfits were previously suggested
- Coordinate multiple AI agents


You need structure. ADK provides that structure.


### The Agent Loop


Every ADK agent follows a loop:


```
1. Receive a message (from user or another agent)
2. Think — the LLM reasons about what to do
3. Act — call a tool, delegate to a sub-agent, or respond
4. Return — send the result back
```


This loop can repeat multiple times within a single request. For example, the stylist agent might:
1. Receive "Style me for a beach vacation"
2. Call `catalog_agent` tool to get the product list
3. Select 3 outfit combinations
4. Call `fitting_tool` for each outfit to generate images
5. Return the structured JSON response


### Core Concepts (With Code From This Repo)


#### LLM Agents


The primary building block. Created with `llmagent.New()`:


```go
// From catalog/agent.go
agent, err := llmagent.New(llmagent.Config{
   Name:        "catalog_agent",                    // Unique identifier
   Model:       m,                                  // Which LLM to use
   Description: "An agent that can search and list products from the catalog",
   Instruction: instructions,                       // Persona prompt (embedded from .md file)
   Tools:       []tool.Tool{listTool, imageTool},   // What the agent can do
})
```


The `Instruction` field is the agent's persona — it tells the LLM who it is and how to behave. In this repo, instructions are written as markdown files and embedded at compile time using Go's `//go:embed` directive:


```go
//go:embed instructions.md
var instructions string
```


This keeps prompts as separate, versionable documents rather than inline strings.


#### Tools


Tools are Go functions that the LLM can call. ADK handles the translation between the LLM's tool-calling format and your typed Go function:


```go
// From catalog/agent.go
type ListProductsArgs struct{}                    // Input (can be empty)
type ListProductsResult struct {
   Products []Product `json:"products"`          // Output
}


func ListProducts(ctx tool.Context, args ListProductsArgs) (ListProductsResult, error) {
   return ListProductsResult{Products: catalogProducts}, nil
}


// Register it:
listTool, _ := functiontool.New(functiontool.Config{
   Name:        "listProducts",
   Description: "list all products in the catalog",
}, ListProducts)
```


ADK automatically generates a JSON schema from your Go structs and sends it to the LLM. When the LLM decides to call `listProducts`, ADK deserializes the arguments, calls your function, and sends the result back.


The `tool.Context` parameter gives tools access to ADK's runtime services — most importantly **artifacts**:


```go
// Save an image as an artifact
ctx.Artifacts().Save(ctx, "my_image", imagePart)


// Load an artifact
resp, _ := ctx.Artifacts().Load(ctx, "my_image")
```


#### Sub-Agent Delegation


An agent can use another agent as a tool via `agenttool.New()`:


```go
// From fittingroom/agent.go
Tools: []tool.Tool{
   loadartifactstool.New(),              // List available artifacts
   imgtool,                              // Get product images
   agenttool.New(catalogAgent, nil),     // Delegate to catalog agent
   fittingTool,                          // Generate try-on image
},
```


When the fitting room agent needs product information, it can call the catalog agent as if it were a regular tool. The LLM sees it in the tool list and can decide to invoke it.


#### Sessions


Sessions track conversation history. ADK's REST API manages them automatically:


```
POST /api/apps/{appName}/users/{userId}/sessions  →  Creates a new session
POST /api/run  (with sessionId)                   →  Runs agent within that session
```


A critical design decision in this app: the **fitting room creates a new session per request** (each try-on is independent), while the **stylist reuses the same session** (so it remembers previous suggestions and can refine based on feedback).


#### State


State is a key-value store attached to a session. Agents read and write state to coordinate:


```go
// Write to state
ctx.State().Set("previously_used_products", "[\"id_bomber\",\"id_hat\"]")


// Read from state
val, err := ctx.State().Get("previously_used_products")
```


The stylist agent uses state to remember which products it already suggested, so it picks different ones next time.


#### Artifacts


Artifacts are named binary objects (usually images) stored per-session. Unlike text responses, they're stored separately and fetched by name:


```go
// Save a generated image as an artifact
artName := fmt.Sprintf("generated_fitting_%s_%s", ctx.InvocationID(), uuid.NewString()[:8])
ctx.Artifacts().Save(ctx, artName, imagePart)


// The frontend fetches it via:
// GET /api/apps/{app}/users/{user}/sessions/{session}/artifacts/{artName}
```


This keeps responses lightweight — the agent returns just the artifact name, and the frontend fetches the binary image data separately.


#### Callbacks


Callbacks are hooks that run at specific points in the agent loop. They can inspect, modify, or short-circuit the execution:






```go
llmagent.Config{
   // Runs before the agent starts — used to save uploaded images
   BeforeAgentCallbacks: []agent.BeforeAgentCallback{SaveIncomingBlobs},


   // Runs before each LLM call — used to inject context
   BeforeModelCallbacks: []llmagent.BeforeModelCallback{
       ExtractAndInjectUserImage,
       InjectPreviousProducts,
   },


   // Runs after each LLM response — used to extract data
   AfterModelCallbacks: []llmagent.AfterModelCallback{SaveSelectedProducts},
}
```


If a callback returns a non-nil response, the default behavior is skipped. For example, a `BeforeModelCallback` that returns a cached response would skip the actual LLM call entirely.


#### JSON Schema Enforcement


Both the fitting room and stylist agents force the LLM to respond in structured JSON:


```go
GenerateContentConfig: &genai.GenerateContentConfig{
   ResponseMIMEType:   "application/json",
   ResponseJsonSchema: fittingSchemaMap(),  // Defines the expected structure
}
```


This ensures the Flutter frontend always receives parseable data, not freeform text.


### The Catalog Agent: Simplest Example


The catalog agent (`catalog/agent.go`) is the simplest agent in the system — a good starting point for understanding ADK patterns.


It has two tools:
1. **`listProducts`** — Returns the full product catalog from a YAML file
2. **`getProductImage`** — Loads a product image from GCS (or local fallback) and saves it as an artifact


The `getProductImage` tool shows an important pattern — **multi-source loading with artifact caching**:


```go
// From tools/imagetool.go — simplified
func GetProductImage(ctx tool.Context, args GetProductImageArgs) (GetProductImageResult, error) {
   // First: check if it's already an artifact
   _, err := ctx.Artifacts().Load(ctx, args.ImageName)
   if err == nil {
       return GetProductImageResult{Image: args.ImageName}, nil  // Cache hit!
   }


   // Second: try loading from GCS bucket
   rc, err := client.Bucket(bucket).Object("catalog-assets/images/" + args.ImageName).NewReader(ctx)
   if err != nil {
       // Third: fall back to local filesystem
       localData, _ := os.ReadFile("../flutter_frontend/assets/images/" + args.ImageName)
       ctx.Artifacts().Save(ctx, args.ImageName, &genai.Part{InlineData: ...})
       return GetProductImageResult{Image: args.ImageName}, nil
   }


   // Save to artifact cache and return
   ctx.Artifacts().Save(ctx, args.ImageName, &genai.Part{InlineData: ...})
   return GetProductImageResult{Image: args.ImageName}, nil
}
```


The tool tries artifacts first, then GCS, then local files. Once loaded, the image is cached as an artifact so subsequent calls are instant.


---


## 🧪 The AI Pipeline: Agents in Action
**Duration: 10 min**


Now let's walk through the two most sophisticated agents — the ones that actually generate images and curate outfits.


### 6.1 The Fitting Room Agent


**File: `adk_backend/fittingroom/agent.go`**


The fitting room agent is the engine behind "Virtual Try-On." When a user uploads their photo and picks a product, this agent generates a composite image of the person wearing that item.


#### The `fitting_tool` — Step by Step


The core logic lives in the `doFitting` function. Here's what happens when the agent calls it:


**Step 1: Resolve the user image**


```go
func doFitting(ctx tool.Context, args FittingToolArgs) (FittingToolResult, error) {
   if len(args.Accessories) > 2 {
       args.Accessories = args.Accessories[:2]  // Safety limit: max 2 items
   }


   var userPart *genai.Part
   if strings.HasPrefix(args.UserImage, "gs://") {
       // If we have a GCS URI from a previous fitting, use it directly
       userPart = &genai.Part{FileData: &genai.FileData{
           FileURI:  args.UserImage,
           MIMEType: gcsURIMimeType(args.UserImage),
       }}
   } else {
       // Otherwise, load the image from artifact storage
       userImgResp, err := ctx.Artifacts().Load(ctx, args.UserImage)
       userPart = userImgResp.Part
   }
```


The user image can come from two sources:
- **An artifact name** (like `upload_abc123_1`) — this is the initial upload, saved by the `SaveIncomingBlobs` callback
- **A `gs://` URI** — this is a previously generated fitting result, stored in GCS for cross-session reuse


This dual-path design is intentional: when the stylist agent later generates outfit try-ons, it reuses the GCS URL from the initial fitting room result so the user's identity stays consistent across all outfits.


**Step 2: Build the multimodal prompt**


```go
   parts := []*genai.Part{
       genai.NewPartFromText(toolInstructions),           // Identity preservation prompt
       genai.NewPartFromText("Reference Person Photo:"),
       userPart,                                          // The user's photo
   }


   for _, acc := range args.Accessories {
       accResp, _ := ctx.Artifacts().Load(ctx, acc)       // Load product image artifact
       parts = append(parts, genai.NewPartFromText("Product Image to Apply:"))
       parts = append(parts, accResp.Part)                // The product photo
   }
```


The `toolInstructions` (embedded from `tool_instructions.md`) is crucial — it tells Gemini to preserve the user's identity (face, body type, skin tone, hair) while applying only the clothing item. Without this prompt engineering, the model might change the person's appearance.


**Step 3: Call Gemini for image generation**


```go
   client, _ := genai.NewClient(ctx, &genai.ClientConfig{
       Backend:  genai.BackendVertexAI,                   // Uses Vertex AI, NOT API key
       Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
       Location: "global",
   })


   resp, _ := client.Models.GenerateContent(ctx, "gemini-2.5-flash-image",
       []*genai.Content{genai.NewContentFromParts(parts, "user")},
       &genai.GenerateContentConfig{
           ResponseModalities: []string{"TEXT", "IMAGE"},  // Request both text and image output
           Temperature:        genai.Ptr(float32(0.2)),    // Low temperature for consistency
       })
```


Notice the authentication difference: the fitting tool uses **Vertex AI** (`genai.BackendVertexAI` with project credentials), while the agents themselves use **API key** auth. This is because `gemini-2.5-flash-image` is called directly through the Vertex AI endpoint for image generation, while the agent orchestration models (`gemini-3-pro-preview`, `gemini-2.5-flash`) use the Gemini API.


**Step 4: Save the result**


```go
   // Find the image in the response (may contain both text + image parts)
   var genPart *genai.Part
   for _, p := range resp.Candidates[0].Content.Parts {
       if p.InlineData != nil {
           genPart = p
           break
       }
   }


   // Save as an ADK artifact
   artName := fmt.Sprintf("generated_fitting_%s_%s", ctx.InvocationID(), uuid.NewString()[:8])
   ctx.Artifacts().Save(ctx, artName, genPart)


   // Also upload to GCS for cross-session reuse
   objectName := fmt.Sprintf("generated-fittings/%s.jpg", ctx.InvocationID())
   w := storageClient.Bucket(bucket).Object(objectName).NewWriter(ctx)
   w.Write(genPart.InlineData.Data)
   gcsURI := fmt.Sprintf("gs://%s/%s", bucket, objectName)


   return FittingToolResult{ArtifactName: artName, GCSUrl: gcsURI}, nil
```


The dual-save (artifact + GCS) is the key to the agent handoff between fitting room and stylist. The artifact provides immediate access within the current session, while the GCS URI allows the stylist (which runs in a different session) to reference the same image later.


#### The `SaveIncomingBlobs` Callback


Before the agent even starts reasoning, this `BeforeAgentCallback` runs to save any images the user uploaded:


```go
func SaveIncomingBlobs(ctx agent.CallbackContext) (*genai.Content, error) {
   for pindex, p := range ctx.UserContent().Parts {
       if p.InlineData != nil {
           aname := fmt.Sprintf("upload_%s_%d", ctx.InvocationID(), pindex)
           ctx.Artifacts().Save(ctx, aname, p)
       }
   }
   return nil, nil  // Return nil to proceed with normal agent execution
}
```


By returning `(nil, nil)`, the callback signals "I'm done preprocessing — now run the agent as normal." If it returned non-nil content, it would short-circuit the agent entirely.


### 6.2 The Stylist Agent


**File: `adk_backend/stylist/agent.go`**


The stylist agent is the most sophisticated in the system. It curates personalized outfit recommendations and supports iterative refinement through conversation.


#### Three Callbacks — The Stylist's Memory


The stylist uses three callbacks to maintain context across multi-turn conversations:


**Callback 1: `InjectPreviousProducts`** (BeforeModel)


The problem: If the user says "show me different options," the LLM might suggest the same products again because it doesn't inherently track what it already recommended.


The solution: After each response, product IDs are saved to session state. Before the next LLM call, this callback reads them and injects a hint:


```go
func InjectPreviousProducts(ctx agent.CallbackContext, req *model.LLMRequest) (*model.LLMResponse, error) {
   prev, err := ctx.State().Get(stateKeyPreviousProducts)  // Read from session state
   if err != nil {
       return nil, nil  // No previous state — first run
   }


   // Append hint to the user's message
   for i := len(req.Contents) - 1; i >= 0; i-- {
       if req.Contents[i].Role == "user" {
           req.Contents[i].Parts = append(req.Contents[i].Parts,
               genai.NewPartFromText(fmt.Sprintf(
                   "IMPORTANT: You previously suggested these products: %s. "+
                   "You MUST pick DIFFERENT complementary products this time.", prev)))
           break
       }
   }
   return nil, nil  // Continue to LLM call
}
```


**Callback 2: `ExtractAndInjectUserImage`** (BeforeModel)


The problem: When the user provides feedback ("make it more casual"), the follow-up message doesn't include the user's photo again. But the fitting tool needs it.


The solution: On the first request, this callback extracts the user image reference and saves it to state. On subsequent requests, it re-injects it:


```go
func ExtractAndInjectUserImage(ctx agent.CallbackContext, req *model.LLMRequest) (*model.LLMResponse, error) {
   var foundImgStr string


   // Search for user image in the latest message
   for i := len(req.Contents) - 1; i >= 0; i-- {
       if req.Contents[i].Role == "user" {
           for _, part := range req.Contents[i].Parts {
               if strings.Contains(part.Text, "User try-on base image") {
                   foundImgStr = part.Text  // Found the GCS URI reference
               }
           }
           break
       }
   }


   if foundImgStr != "" {
       ctx.State().Set(stateKeyUserImageStr, foundImgStr)  // Save for later
   } else {
       // Not in current message — retrieve from state and inject
       val, _ := ctx.State().Get(stateKeyUserImageStr)
       if savedImgStr, ok := val.(string); ok {
           // Inject into the latest user message
           req.Contents[last].Parts = append(req.Contents[last].Parts,
               genai.NewPartFromText("REMINDER: Use this image: " + savedImgStr))
       }
   }
   return nil, nil
}
```


**Callback 3: `SaveSelectedProducts`** (AfterModel)


After the LLM responds with outfit suggestions, this callback parses the JSON to extract product IDs and saves them for the `InjectPreviousProducts` callback to use next time:


```go
func SaveSelectedProducts(ctx agent.CallbackContext, resp *model.LLMResponse, respErr error) (*model.LLMResponse, error) {
   for _, part := range resp.Content.Parts {
       ids := extractProductIDs(part.Text)  // Parse JSON → extract product IDs
       if len(ids) > 0 {
           data, _ := json.Marshal(ids)
           ctx.State().Set(stateKeyPreviousProducts, string(data))  // Save to state
       }
   }
   return nil, nil  // Don't modify the response
}
```


Together, these three callbacks create a feedback loop:


```
Request 1:  User sends styling request + user image
           → ExtractAndInjectUserImage SAVES image to state
           → LLM generates 3 outfits
           → SaveSelectedProducts SAVES product IDs to state


Request 2:  User says "make it more casual"
           → ExtractAndInjectUserImage INJECTS saved image into prompt
           → InjectPreviousProducts INJECTS "don't reuse these IDs"
           → LLM generates 3 NEW outfits
           → SaveSelectedProducts UPDATES product IDs in state
```


### 6.3 The Root Agent


**File: `adk_backend/rootagent/agent.go`**


The simplest agent — just 31 lines:


```go
func NewRootAgent(apiKey string, fittingAgent, catalogAgent, stylistAgent agent.Agent) (agent.Agent, error) {
   m, _ := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{APIKey: apiKey})


   return llmagent.New(llmagent.Config{
       Name:        "root_agent",
       Model:       m,
       Description: "A root agent that delegates to other agents",
       Instruction: "You are a helpful shopping assistant. If the user asks about fitting " +
           "items or generating images, delegate to the fitting room agent. If the user " +
           "asks about products, delegate to the catalog agent. If the user asks for " +
           "styling advice, delegate to the stylist agent.",
       SubAgents: []agent.Agent{fittingAgent, catalogAgent, stylistAgent},
   })
}
```


It uses `gemini-2.5-flash` (the fastest model) because routing decisions are simple — the LLM just needs to read the user's intent and pick the right sub-agent. No tools needed; `SubAgents` handles delegation automatically.


---


## 📱 Flutter Frontend Architecture
**Duration: 10 min**


The Flutter frontend is a fully functional retail shopping app. The AI features live in `flutter_frontend/lib/workshop_tasks/`, separate from the pre-built shopping experience in `core_app/`.


### The MVVM Pattern


The app follows **Model-View-ViewModel** architecture with the Provider package:


```
┌──────────────────┐    ┌────────────────────┐    ┌──────────────────┐
│  View (Widget)   │◀───│ ViewModel (Provider)│◀───│  Service (HTTP)  │
│                  │    │                     │    │                  │
│ • Renders UI     │    │ • Holds state       │    │ • Makes API calls│
│ • User gestures  │───▶│ • Business logic    │───▶│ • Parses response│
│ • Listens for    │    │ • notifyListeners() │    │ • Returns data   │
│   state changes  │    │                     │    │                  │
└──────────────────┘    └────────────────────┘    └──────────────────┘
```


Each layer has a clear role:


- **Model**: Data classes like `Product`, `Outfit`, `StyleRequest`, and enums like `TryOnState`
- **ViewModel** (`ChangeNotifier`): Holds the current state and broadcasts changes to the UI via `notifyListeners()`
- **View** (Widget): Subscribes to the ViewModel with `context.watch<Provider>()` and rebuilds when state changes
- **Service**: Makes HTTP calls to the ADK backend and returns typed data


### The Service Layer


Services are defined as abstract interfaces, with ADK-specific implementations:


```dart
// Abstract interface — defines WHAT the service does
abstract class TryItOnService {
 Future<(Uint8List?, String?)> generateTryOnImage(
   Uint8List userImageBytes,
   Uint8List productImageBytes,
 );
}


// Concrete implementation — defines HOW (via ADK REST API)
class AdkFittingRoomService implements TryItOnService { ... }
```


This separation means you could swap the ADK backend for Firebase AI, a mock service, or any other implementation without changing the rest of the app.


### The 3-Step API Pattern


Both `AdkFittingRoomService` and `AdkStylingService` follow the same pattern to talk to the ADK backend:


**Step 1: Create a session**


```dart
Future<String> _createSession() async {
 final url = Uri.parse(
   '$_baseUrl/apps/${Uri.encodeComponent(_appName)}/users/$_userId/sessions');
 final response = await _client.post(url,
   headers: {'Content-Type': 'application/json'},
   body: jsonEncode({}));
 final body = jsonDecode(response.body) as Map<String, dynamic>;
 return body['id'] as String;  // Returns the session ID
}
```


**Step 2: Run the agent**


```dart
Future<(String?, String?)> _runAgent({required String sessionId, ...}) async {
 final requestBody = jsonEncode({
   'appName': _appName,
   'userId': _userId,
   'sessionId': sessionId,
   'newMessage': {
     'role': 'user',
     'parts': [
       {'text': 'Generate a virtual try-on...'},
       {'inlineData': {'mimeType': 'image/jpeg', 'data': base64Encode(userImageBytes)}},
       {'inlineData': {'mimeType': 'image/png', 'data': base64Encode(productImageBytes)}},
     ],
   },
 });
 final response = await _client.post(Uri.parse('$_baseUrl/run'), body: requestBody);
 // Parse response events for artifact name and GCS URL...
}
```


**Step 3: Fetch the artifact**


```dart
Future<Uint8List?> _loadArtifact({required String sessionId, required String artifactName}) async {
 final url = Uri.parse(
   '$_baseUrl/apps/$_appName/users/$_userId/sessions/$sessionId/artifacts/$artifactName');
 final response = await _client.get(url);
 final part = jsonDecode(response.body) as Map<String, dynamic>;
 final data = part['inlineData']['data'] as String;
 return base64Decode(data);  // Returns raw image bytes
}
```


A critical design difference: the **fitting room service creates a new session for every request** (`_createSession()` is called each time), while the **styling service reuses the same session** (`_sessionId ??= await _createSession()`) to enable multi-turn conversation.


### State Management: The TryItOnProvider


**File: `workshop_tasks/step_1_try_it_on/providers/try_it_on_provider.dart`**


The `TryItOnProvider` manages the entire try-on flow. It uses a `TryOnState` enum as a state machine:


```dart
enum TryOnState { initial, imagePicked, generating, success, error }


class TryItOnProvider with ChangeNotifier {
 TryOnState _state = TryOnState.initial;
 Uint8List? _userImageBytes;
 Uint8List? _generatedImage;
 String? _errorMessage;
```


Private state transitions ensure consistency — you never update the state without also clearing stale data and notifying the UI:


```dart
 void _setGenerating() {
   _state = TryOnState.generating;
   _errorMessage = null;            // Clear any previous error
   _wasLastGenerationCached = false;
   notifyListeners();               // Tell the UI to rebuild
 }


 void _setSuccess(Uint8List image, {bool isCached = false}) {
   _generatedImage = image;
   _errorMessage = null;
   _wasLastGenerationCached = isCached;
   _state = TryOnState.success;
   notifyListeners();
 }
```


The main generation method ties it all together:


```dart
 Future<String?> generateTryOnImage(String productImagePath) async {
   final userImageBytes = _userImageBytes;
   if (userImageBytes == null) {
     _setError('No image selected.');
     return _errorMessage;
   }


   // Check local cache first
   final cachedImage = ImageUtils.getCachedImage(_sessionCache, userImageBytes, productUint8List);
   if (cachedImage != null) {
     _setSuccess(cachedImage, isCached: true);
     return null;
   }


   _setGenerating();  // Triggers loading state in UI


   try {
     final (generatedBytes, gcsUrl) = await _aiService.generateTryOnImage(
       userImageBytes, productUint8List);
     if (generatedBytes != null) {
       _fittingGcsUrl = gcsUrl;     // Save for the stylist agent later
       _setSuccess(generatedBytes);
     }
   } catch (e) {
     _setError(e.toString());
   }
   return _state == TryOnState.success ? null : _errorMessage;
 }
```


### The UI: Screens as a State Router


**File: `workshop_tasks/step_1_try_it_on/ui/2_try_it_on_screen.dart`**


The try-on screen uses Dart 3's pattern matching with `AnimatedSwitcher` to route between sub-screens based on the provider's state:


```dart
class TryItOnScreen extends StatelessWidget {
 @override
 Widget build(BuildContext context) {
   final tryOnProvider = context.watch<TryItOnProvider>();


   return ScaffoldWithBackgroundNoise(
     appBar: const TryOnAppBar(),
     body: AnimatedSwitcher(
       duration: AppDurations.medium,
       child: switch (tryOnProvider.state) {
         TryOnState.initial || TryOnState.error => const ChooseImageScreen(),
         TryOnState.imagePicked || TryOnState.generating => LoadingScreen(
           userImage: tryOnProvider.userImageBytes),
         TryOnState.success => const FittingRoomScreen(),
       },
     ),
   );
 }
}
```


`context.watch<TryItOnProvider>()` subscribes to the provider. Whenever `notifyListeners()` is called, this widget rebuilds, and `AnimatedSwitcher` smoothly transitions between screens. There's no `Navigator.push` — the screen content changes in-place based on the state enum.


### The Agentic Handoff: Fitting Room → Stylist


The most interesting UX pattern is how the app passes context from the fitting room agent to the stylist agent.


In `5_fitting_room.dart`, after the try-on image is generated, the "Style Me" button opens a form. When the user submits:


```dart
// From 1_style_me_form_sheet.dart
Navigator.pop(context, StyleRequest(
 location: _locationController.text.trim(),
 occasion: _occasionController.text.trim(),
 notes: _notesController.text.trim(),
 gcsUserImageUrl: provider.fittingGcsUrl,          // GCS URI from fitting result
 userImageData: provider.fittingGcsUrl == null
     ? provider.userImageBytes : null,              // Fallback to raw bytes
 selectedProductId: provider.selectedProduct?.id,   // Product they already tried on
 selectedProductTitle: provider.selectedProduct?.title,
));
```


The `StyleRequest` bundles everything the stylist needs:
- **Location and occasion** — text context for styling
- **GCS user image URL** — so the stylist can reuse the exact same user representation
- **Selected product** — so the stylist includes it in every outfit


This is the **agentic handoff** — seamlessly transferring multimodal context from one AI agent to another, with the user only seeing a simple form.


### The Styling Flow: StylingProvider


**File: `workshop_tasks/step_2_style_me/providers/styling_provider.dart`**


The `StylingProvider` is simpler than `TryItOnProvider` because it delegates most complexity to the backend:


```dart
class StylingProvider with ChangeNotifier {
 StylingState _state = StylingState.initial;
 List<Outfit> _outfits = [];


 Future<bool> getStyleSuggestions(StyleRequest request) async {
   if (_state == StylingState.loading) return false;  // Prevent spam
   _currentRequest = request;
   _setLoading();
   try {
     final suggestions = await _stylingService.getStyleSuggestions(request);
     _setSuccess(suggestions);
     return true;
   } catch (e) {
     _setError('Something went wrong.');
   }
   return false;
 }


 // Multi-turn refinement — uses the same session
 Future<bool> refineWithFeedback(String feedback) async {
   if (_state == StylingState.loading) return false;
   _setLoading();
   try {
     final suggestions = await _stylingService.refineWithFeedback(feedback);
     _setSuccess(suggestions);
     return true;
   } catch (e) {
     _setError('Something went wrong.');
   }
   return false;
 }
}
```


The `refineWithFeedback` method sends a plain text message to the same session — the backend's `InjectPreviousProducts` and `ExtractAndInjectUserImage` callbacks handle all the context management automatically.


---


## 🚀 Run the App Locally
**Duration: 3 min**


### Start the Backend


Open a terminal and start the ADK Go backend:


```bash
cd ~/thread-count-workshop/adk_backend
./run.sh
```


You should see:


```
Starting agent server on port 8080 ...
```


The server exposes:
- **REST API** at `http://localhost:8080/api/` — used by the Flutter app
- **ADK Dev UI** at `http://localhost:8080/` — interactive chat for testing agents directly


### Start the Flutter Frontend


Open a **second terminal** and start the Flutter app:


```bash
cd ~/thread-count-workshop/flutter_frontend
flutter pub get
flutter run -d web-server --web-port=61983 --web-hostname=0.0.0.0
```


> aside positive
> In Cloud Shell, use `flutter run -d web-server` instead of `-d chrome` since there's no local browser. Cloud Shell provides a web preview URL.


### Verify the Connection


1. Open the Flutter app in your browser (via Cloud Shell's Web Preview on port 8888)
2. Browse the product catalog and select an item
3. Tap the person icon (👤) to start the Try-On flow
4. Upload a photo and watch the AI generate a try-on image
5. Tap "Style Me" to get outfit recommendations


### Explore the ADK Dev UI


The ADK Dev UI (port 8080) lets you chat with agents directly — useful for debugging:
- Select the `fitting room` app
- Send a text message to test the agent without the Flutter UI
- Inspect session state, artifacts, and the full conversation history


---


## ☁️ Deploy to Cloud Run
**Duration: 5 min**


### Deploy the Backend


From the `adk_backend` directory:


```bash
cd ~/thread-count-workshop/adk_backend


gcloud run deploy thread-count-backend \
 --source . \
 --region us-central1 \
 --allow-unauthenticated \
 --set-env-vars "GEMINI_API_KEY=$(grep GEMINI_API_KEY .env | cut -d= -f2),GOOGLE_CLOUD_PROJECT=$PROJECT_ID,GCS_BUCKET=thread-count-$PROJECT_ID" \
 --memory 1Gi \
 --cpu 2 \
 --timeout 300s \
 --min-instances 0 \
 --max-instances 3
```


When the deploy finishes, copy the **Service URL** (e.g., `https://thread-count-backend-xyz-uc.a.run.app`).


### Update the Frontend Configuration


Edit `flutter_frontend/lib/app_config.dart` to point to the Cloud Run URL:


```dart
class AppConfig {
 static const String adkBackendUrl =
     'https://thread-count-backend-xyz-uc.a.run.app/api';  // ← Your Cloud Run URL
}
```


### Deploy the Frontend (Optional)


Build the Flutter web app and deploy it to Cloud Run as well:


```bash
cd ~/thread-count-workshop/flutter_frontend
flutter build web
```


Then deploy the `build/web` directory using any static hosting service or Cloud Run with a simple Dockerfile.


### Verify the Deployment


Open the Cloud Run URL in your browser and run through the full flow:
1. **Browse** → Select a product
2. **Try On** → Upload your photo → See the AI-generated image
3. **Style Me** → Fill out location/occasion → See curated outfits
4. **Feedback** → Type "make it more casual" → See updated outfits
5. **Add to Bag** → Complete the shopping flow


---


## 🎉 Conclusion
**Duration: 2 min**


### What You Built


You've explored a complete AI-powered retail experience with:


- ✅ A **multi-agent backend** with 4 specialized agents working together
- ✅ A **virtual fitting room** that generates personalized try-on images
- ✅ An **AI stylist** that curates outfits and refines them through conversation
- ✅ A **cross-platform Flutter app** that connects to the agent backend
- ✅ **Cloud Run deployment** for scalable, serverless hosting


### Key Concepts


| Concept | Where You Saw It |
|---------|-----------------|
| **ADK multi-agent orchestration** | Root agent routing to fitting room, catalog, and stylist agents |
| **Gemini multimodal image generation** | `fitting_tool` combining user photos with product images |
| **Session state for conversational AI** | Stylist reusing sessions for iterative feedback |
| **Artifact storage for binary data** | Separating image storage from text responses |
| **Callbacks for middleware logic** | `SaveIncomingBlobs`, `InjectPreviousProducts`, `SaveSelectedProducts` |
| **MVVM + Provider in Flutter** | `TryItOnProvider` and `StylingProvider` with `ChangeNotifier` |
| **Agentic handoff** | `StyleRequest` passing multimodal context between agents |


### Next Steps


- 🎨 **Customize agent prompts** — edit `instructions.md` to change the stylist's personality
- 🛍️ **Add more products** — update `catalog.yaml` with new items
- 📱 **Build for mobile** — run `flutter build ios` or `flutter build apk`
- 🔄 **Add persistent sessions** — replace `InMemoryService` with a database-backed implementation
- 🔒 **Add authentication** — secure the Cloud Run endpoint with IAM


### Resources


- [ADK Documentation](https://adk.dev) — Official Agent Development Kit docs
- [ADK Go Source Code](https://github.com/google/adk-go) — GitHub repository
- [ADK Go Package Reference](https://pkg.go.dev/google.golang.org/adk) — API reference
- [Gemini API Documentation](https://ai.google.dev/gemini-api/docs) — Model capabilities and guides
- [Flutter Provider Package](https://pub.dev/packages/provider) — State management docs
- [Cloud Run Documentation](https://cloud.google.com/run/docs) — Deployment and scaling guides



