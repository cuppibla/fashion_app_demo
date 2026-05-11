# 👗 Build a Virtual Fitting Room & AI Stylist with Flutter, ADK Go & Gemini


---


## Introduction
**Duration: 3 min**


### What You'll Build


In this codelab, you'll step into the shoes of a developer building **Fashion App**, a Flutter shopping app for a fictional retail brand. Your mission: add two AI-powered features that transform the online shopping experience.


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
| **Agent Reasoning (Pro)** | Gemini 3.1 Pro Preview | Powers the fitting room and stylist agents |
| **Agent Reasoning (Flash)** | Gemini 3 Flash Preview | Powers the root and catalog agents (lightweight routing/lookup) |
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


### 2. Set Up Flutter SDK


Cloud Shell ships with Flutter pre-installed at `/google/flutter`. Because that directory is owned by a different system user, you'll hit a `fatal: detected dubious ownership` error the first time you run `flutter`. Add it to git's safe-directory list once:


```bash
git config --global --add safe.directory /google/flutter
```


Verify Flutter is on your `PATH` and working:


```bash
flutter --version
```


The first run downloads the Dart SDK and builds the Flutter tool — give it a minute. You should see something like `Flutter 3.x • channel stable`.


> aside positive
> **Prefer the bundled Flutter.** If you'd rather pin to your own stable clone (e.g., for a specific Flutter version), run:
> ```bash
> cd ~
> git clone https://github.com/flutter/flutter.git -b stable --depth 1
> echo 'export PATH="$HOME/flutter/bin:$PATH"' >> ~/.bashrc
> source ~/.bashrc
> ```
> Note the `$HOME/flutter/bin` comes **before** `$PATH` so your clone wins over `/google/flutter`.


### 3. Clone the Repository


```bash
cd ~
git clone https://github.com/cuppibla/fashion_app_demo.git
cd fashion_app_demo
```


### 4. Explore the Project Structure


```
fashion_app_demo/
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
gcloud projects create fashion-app-demo --name="Fashion App Demo"
gcloud config set project fashion-app-demo
```


> aside positive
> If the project ID is taken, append a random number (e.g., `fashion-app-demo-42`).


### 2. Link a Billing Account


> aside negative
> **This codelab assumes you already have an active billing account.** Vertex AI, Cloud Build, and Cloud Run all require billing — without it, the next step (`gcloud services enable`) will fail for `aiplatform.googleapis.com` and `cloudbuild.googleapis.com`, and even the APIs that do enable won't actually run workloads.


List your billing accounts:


```bash
gcloud billing accounts list
```


**Look at the `OPEN` column.** It must say `True`. If it says `False` (common with an expired free trial), the account is closed and won't actually pay for anything — skip ahead to the troubleshooting block below before continuing.


Copy the `ACCOUNT_ID` of an `OPEN: True` account (looks like `0X0X0X-0X0X0X-0X0X0X`) and link it to your project:


```bash
gcloud billing projects link fashion-app-demo \
 --billing-account=YOUR_BILLING_ACCOUNT_ID
```


Verify the link:


```bash
gcloud billing projects describe fashion-app-demo
```


You should see `billingEnabled: true`. If you see `billingEnabled: false` even after linking, the account is closed (`OPEN: False`) — see the troubleshooting block below.


> aside positive
> **Don't have a billing account?** Create one at [console.cloud.google.com/billing](https://console.cloud.google.com/billing). New Google Cloud users get **$300 in free credits**, which is plenty for this codelab — the Gemini API calls and Cloud Run usage here cost only a few cents. You'll need a credit card for verification, but you won't be charged unless you exceed the free tier and opt in to a paid account.


> aside negative
> **Troubleshooting: `OPEN: False` / `billingEnabled: false` / `UREQ_PROJECT_BILLING_NOT_OPEN`.**
>
> If you ran `gcloud services enable ...` and got an error like:
> ```
> ERROR: (gcloud.services.enable) FAILED_PRECONDITION: Billing account for
> project '...' is not open. Billing must be enabled for activation of
> service(s) '...' to proceed.
> ...
> reason: UREQ_PROJECT_BILLING_NOT_OPEN
> ```
> **that means you need a payment method on the billing account.** Your free-trial account has closed (trial expired, credit used up, or no card on file), so even though the project is "linked" to it, no service can be turned on. Fix it like this:
>
> 1. Open [console.cloud.google.com/billing](https://console.cloud.google.com/billing) in your browser.
> 2. Click on the trial account (the one with `OPEN: False`).
> 3. Look for **Upgrade**, **Reactivate**, or **Add payment method** — the exact button depends on the trial's state. Add a credit/debit card.
> 4. Alternatively, click **Create account** at the top of the billing page to set up a fresh paid account with a new card.
> 5. Confirm the fix back in Cloud Shell:
>    ```bash
>    gcloud billing accounts list --filter="open:true"
>    ```
>    This should now return at least one account. Re-run the `link` command above with that `ACCOUNT_ID`, then retry the `enable` command in the next step.
>
> The codelab's actual usage costs only a few cents — adding a card unlocks the APIs without committing you to ongoing charges.


### 3. Enable Required APIs


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


### 4. Create a GCS Bucket


```bash
export PROJECT_ID=$(gcloud config get-value project)
gcloud storage buckets create gs://fashion-app-$PROJECT_ID \
 --location=us-central1 \
 --uniform-bucket-level-access
```


> aside positive
> Bucket names are globally unique across all of Google Cloud. Since `$PROJECT_ID` is already unique, `fashion-app-$PROJECT_ID` will be too. The `--uniform-bucket-level-access` flag is the modern default — it disables legacy ACLs in favor of IAM.


### 5. Upload Product Catalog Images


The backend's `getProductImage` tool reads from `gs://$GCS_BUCKET/catalog-assets/images/<image_name>`. Upload the catalog images to that exact path:


```bash
cd ~/fashion_app_demo
gcloud storage cp flutter_frontend/assets/images/*.png \
 gs://fashion-app-$PROJECT_ID/catalog-assets/images/
```


Verify the upload (you should see a list of `.png` files):


```bash
gcloud storage ls gs://fashion-app-$PROJECT_ID/catalog-assets/images/
```


### 6. Configure the `.env` File


```bash
cd ~/fashion_app_demo/adk_backend
cat > .env << EOF
GOOGLE_CLOUD_PROJECT=$PROJECT_ID
GCS_BUCKET=fashion-app-$PROJECT_ID
EOF
```


> aside positive
> No API key is needed. Both Vertex AI (Gemini calls) and Cloud Storage (image uploads/downloads) authenticate through **Application Default Credentials** — set up in the next step. The `.env` file just tells the backend which project and bucket to use.


### 7. Authenticate with Application Default Credentials


**You must run this before starting the backend locally.** The Go backend uses ADC to authenticate every call to Vertex AI (Gemini) and Cloud Storage. Without ADC, the backend will start up but every try-on request will fail with a 401 `CREDENTIALS_MISSING`.


One credential covers both services. Run these two commands in order:


```bash
# 1. Log in (opens a browser; in Cloud Shell, paste the verification code back)
gcloud auth application-default login

# 2. Attach your project as the quota / billing project for ADC
gcloud auth application-default set-quota-project $(gcloud config get-value project)
```


Verify ADC is healthy:


```bash
gcloud auth application-default print-access-token | head -c 20 && echo "..."
```


You should see ~20 characters of a token followed by `...`. If it errors, the login didn't take — re-run step 1.


> aside positive
> **One auth path for everything.** The Go code uses `Backend: genai.BackendVertexAI` with your project ID for Gemini, and the Cloud Storage client uses the same ADC for `gs://` reads/writes. When you deploy to Cloud Run later, ADC is replaced by the Cloud Run service account automatically — no auth code changes needed.


> aside negative
> **You'll need to redo this for each new account or fresh Cloud Shell session.** ADC is stored in `~/.config/gcloud/application_default_credentials.json` per Google account. If you switch accounts (or Cloud Shell hands you an ephemeral home), the file may be missing or out of date. Run both commands above again. Symptoms of stale ADC: backend logs show `401 CREDENTIALS_MISSING` even after the codelab's setup was completed earlier.


---


## 🏗️ Architecture Overview
**Duration: 5 min**


Now that the environment is ready, let's understand how the system works before looking at the code.


### The Four-Agent System


The backend is built as a **multi-agent system** using ADK (Agent Development Kit) for Go. Four agents work together, each with a specific responsibility:


```
                   ┌──────────────┐
                   │ Root Agent   │  ← Routes requests to the right agent
                   │ (gemini-3-   │
                   │  flash-      │
                   │  preview)    │
                   └──────┬───────┘
                          │
           ┌──────────────┼──────────────┐
           ▼              ▼              ▼
  ┌────────────────┐ ┌──────────┐ ┌────────────────┐
  │ Fitting Room   │ │ Catalog  │ │ Stylist        │
  │ Agent          │ │ Agent    │ │ Agent          │
  │ (gemini-3.1-   │ │ (gemini- │ │ (gemini-3.1-   │
  │  pro-preview)  │ │  3-flash-│ │  pro-preview)  │
  │                │ │  preview)│ │                │
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
| **Root Agent** | `gemini-3-flash-preview` | Traffic cop. Reads the user's message and delegates to the right specialist agent. Uses a fast, lightweight model because it only needs to make routing decisions. |
| **Catalog Agent** | `gemini-3-flash-preview` | Product expert. Loads the product catalog from a YAML file and answers product queries. Also lightweight — it's just looking up data. |
| **Fitting Room Agent** | `gemini-3.1-pro-preview` | Virtual try-on specialist. Takes a user photo + product image and generates a composite image of the person wearing that item. Uses a more capable model because it needs to reason about images. |
| **Stylist Agent** | `gemini-3.1-pro-preview` | Fashion advisor. Given location, occasion, and preferences, it curates 3 outfit combinations from the catalog. Can generate try-on images for each outfit. Also uses the capable model for creative reasoning. |


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
       Backend:  genai.BackendVertexAI,                 // Vertex AI endpoint
       Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),     // From your .env
       Location: "global",                              // Multi-region endpoint
   })


   resp, _ := client.Models.GenerateContent(ctx, "gemini-2.5-flash-image",
       []*genai.Content{genai.NewContentFromParts(parts, "user")},
       &genai.GenerateContentConfig{
           ResponseModalities: []string{"TEXT", "IMAGE"},  // Request both text and image output
           Temperature:        genai.Ptr(float32(0.2)),    // Low temperature for consistency
       })
```


All four agents and the image-gen tool share a single authentication path: `Backend: genai.BackendVertexAI` with the project ID, authenticated via Application Default Credentials. The orchestration models (`gemini-3.1-pro-preview`, `gemini-3-flash-preview`) and the image model (`gemini-2.5-flash-image`) all sit behind the same Vertex AI endpoint, and the same ADC also authorizes Cloud Storage access — one credential, every call.


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
func NewRootAgent(project string, fittingAgent, catalogAgent, stylistAgent agent.Agent) (agent.Agent, error) {
   m, _ := gemini.NewModel(ctx, "gemini-3-flash-preview", &genai.ClientConfig{
       Backend:  genai.BackendVertexAI,
       Project:  project,
       Location: "global",
   })


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


It uses `gemini-3-flash-preview` (the fastest model) because routing decisions are simple — the LLM just needs to read the user's intent and pick the right sub-agent. No tools needed; `SubAgents` handles delegation automatically.


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
**Duration: 5 min**


For a smooth Cloud Shell experience, the Go backend serves the compiled Flutter web app from the **same port** (8080). One process, one preview URL, no cross-origin headaches, no editing of config files.


### Before you start — sanity-check ADC


The backend needs **Application Default Credentials** to call Vertex AI. If you finished step 7 of the project setup in *this* Cloud Shell session and *this* Google account, you're good. If you're returning after a break, switched accounts, or aren't sure, take 5 seconds to verify:


```bash
gcloud auth application-default print-access-token | head -c 20 && echo "..."
```


If that prints ~20 characters of a token, you're set. If it errors, **re-run step 7 of the project setup**:
```bash
gcloud auth application-default login
gcloud auth application-default set-quota-project $(gcloud config get-value project)
```


You'll use **two Cloud Shell terminals**:
- **Terminal A** — runs the backend continuously (`./run.sh`). Leave it open.
- **Terminal B** — runs the Flutter web build once (`flutter build web`). Exits when done.

The order doesn't matter — you can start either first. But for the cleanest first-run experience, build Flutter first so the backend has a UI to serve from the moment it starts.


### 1. Terminal B — Build the Flutter Web Bundle (one-shot)


Open a new Cloud Shell tab (the **+** at the top of the terminal panel), then:


```bash
cd ~/fashion_app_demo/flutter_frontend
flutter pub get
flutter build web
```


This produces `flutter_frontend/build/web/` — a directory of static files (HTML, JS, assets) — and exits when finished. The backend will serve these as soon as it sees the directory exists.


> aside positive
> Re-run `flutter build web` in Terminal B whenever you change Dart code. The backend in Terminal A picks up the new build on the next request — **no restart needed**.


> aside negative
> **`fatal: detected dubious ownership in repository at '/google/flutter'` (×4)**
>
> You'll see this if you're in a fresh Cloud Shell session or signed in with a different Google account than when you ran the setup. The fix is the same one-liner from [§ Prerequisites step 2.2](#2-set-up-flutter-sdk) — it just needs to run once per account:
> ```bash
> git config --global --add safe.directory /google/flutter
> ```
> Then re-run `flutter pub get && flutter build web`. The config is stored in `~/.gitconfig`, so it persists across future sessions for **this** account.


### 2. Terminal A — Start the Backend (long-running)


In your original Cloud Shell terminal:


```bash
cd ~/fashion_app_demo/adk_backend
./run.sh
```


You should see something like:


```
Serving Flutter web build from ../flutter_frontend/build/web
```


**Leave this terminal running** — the backend stays up for as long as `run.sh` is alive. To stop it, hit `Ctrl+C`.

The server exposes everything on port 8080:
- **`/`** — Flutter web app (the shopping UI)
- **`/api/`** — ADK REST endpoints (called by the Flutter app)
- **ADK Dev UI** — also at `/` when there's no Flutter build; useful for direct agent debugging


> aside negative
> **Troubleshooting `./run.sh` output**
>
> **`Flutter build not found — run flutter build web ...`**
> The backend started, but there's no UI to serve. You skipped step 1, or the build failed. In Terminal B:
> ```bash
> cd ~/fashion_app_demo/flutter_frontend
> flutter pub get
> flutter build web
> ```
> Leave Terminal A's backend running — it picks up the new build on the next request, no restart needed.
>
> **`Error 401 ... CREDENTIALS_MISSING` when the user uploads a photo**
> The backend is running but can't authenticate to Vertex AI. ADC is missing or stale. From Terminal A, redo step 7 of the project setup:
> ```bash
> gcloud auth application-default login
> gcloud auth application-default set-quota-project $(gcloud config get-value project)
> ```
> Then in Terminal A, also clear any stale credential env vars before restarting:
> ```bash
> unset GOOGLE_APPLICATION_CREDENTIALS
> ./run.sh
> ```
> Sanity-check that ADC works for Vertex AI specifically (replace `$PROJECT_ID` if needed):
> ```bash
> ACCESS_TOKEN=$(gcloud auth application-default print-access-token)
> curl -sS -X POST \
>  -H "Authorization: Bearer $ACCESS_TOKEN" \
>  -H "Content-Type: application/json" \
>  -H "x-goog-user-project: $(gcloud config get-value project)" \
>  "https://aiplatform.googleapis.com/v1beta1/projects/$(gcloud config get-value project)/locations/global/publishers/google/models/gemini-3-flash-preview:generateContent" \
>  -d '{"contents":[{"role":"user","parts":[{"text":"ping"}]}]}'
> ```
> If this returns a JSON `candidates` array, ADC is healthy; the issue is elsewhere. If it 403s with `SERVICE_DISABLED`, run `gcloud services enable aiplatform.googleapis.com` again. If it 403s with `PERMISSION_DENIED`, your account lacks `roles/aiplatform.user` on this project.
>
> **`Warning: The user provided project/location will take precedence over the API key from the environment variable.`**
> You'll see this once per agent (four times total). It's harmless — the Go code is intentionally using ADC + project, and the SDK is just noting that any `GOOGLE_API_KEY` in your environment is being ignored. The most common cause is a stale `.env` file from an older version of the codelab that included a `GOOGLE_API_KEY=...` line. To silence the warnings, regenerate `.env`:
> ```bash
> export PROJECT_ID=$(gcloud config get-value project)
> cd ~/fashion_app_demo/adk_backend
> cat > .env << EOF
> GOOGLE_CLOUD_PROJECT=$PROJECT_ID
> GCS_BUCKET=fashion-app-$PROJECT_ID
> EOF
> ```
> Then restart `./run.sh` in Terminal A. **Also check that `$PROJECT_ID` actually has a value** before generating `.env` — `gcloud config get-value project` returns `(unset)` if the current gcloud config has no project, which produces an empty `.env` and a `bucket doesn't exist` error.
>
> **`storage: bucket doesn't exist: Error 404`**
> Your `.env` has an empty or wrong `GCS_BUCKET`. Run `cat ~/fashion_app_demo/adk_backend/.env` — if `GCS_BUCKET=fashion-app-` (truncated, no project), `gcloud config get-value project` returned empty when you ran step 6. Fix:
> ```bash
> gcloud config set project YOUR_PROJECT_ID
> # then regenerate .env as in step 6
> ```
> Also confirm the bucket actually exists on this project: `gcloud storage ls gs://fashion-app-$(gcloud config get-value project)`. If not, run step 4 (create bucket) and step 5 (upload images).


### 3. Open Web Preview


1. In Cloud Shell, click the **Web Preview** icon (top-right) → **Preview on port 8080**
2. The Flutter shopping app loads in a new tab
3. Browse the product catalog and select an item
4. Tap the person icon (👤) to start the Try-On flow
5. Upload a photo and watch the AI generate a try-on image
6. Tap "Style Me" to get outfit recommendations
7. Type follow-up feedback like "make it more casual" — same-session refinement


> aside negative
> **Why same-origin matters in Cloud Shell.** If the backend ran on 8080 and Flutter ran on 8081 separately, the Flutter JS (running in your browser) would try to call `http://localhost:8080/api` — but `localhost` from the browser means *your laptop*, not the Cloud Shell VM. Serving both from the same port fixes this entirely: the Flutter app's API calls go to `/api` relative to the same preview URL.


---


## ☁️ Deploy to Cloud Run
**Duration: 5 min**


### Bundle the Flutter Build into the Backend


The Cloud Run container ships both the API and the UI from one image. Copy the Flutter web build into `adk_backend/flutter_web/` — that's the first path the Go server checks when picking which UI to serve:


```bash
cd ~/fashion_app_demo/flutter_frontend
flutter build web
rm -rf ../adk_backend/flutter_web
cp -r build/web ../adk_backend/flutter_web
```


(If you've been iterating locally, you may already have `build/web` from the Run-Locally step. Re-running `flutter build web` is still fine.)


### Deploy the Backend (Serves API + UI)


```bash
cd ~/fashion_app_demo/adk_backend

gcloud run deploy fashion-app-backend \
 --source . \
 --region us-central1 \
 --allow-unauthenticated \
 --set-env-vars "GOOGLE_CLOUD_PROJECT=$PROJECT_ID,GCS_BUCKET=fashion-app-$PROJECT_ID" \
 --memory 1Gi \
 --cpu 2 \
 --timeout 300s \
 --min-instances 0 \
 --max-instances 3
```


When the deploy finishes, you'll get a **Service URL** like `https://fashion-app-backend-xyz-uc.a.run.app`. Open it in a browser — the Flutter shopping app loads from `/`, and its API calls go to `/api/` on the same host. **No frontend config edits needed, no API key passed.**


> aside positive
> Because [app_config.dart](flutter_frontend/lib/app_config.dart) uses the same-origin relative URL `/api`, the deployed Flutter app automatically talks to the deployed backend. The same build works locally and on Cloud Run with zero changes.


> aside negative
> **Cloud Run service account permissions.** The default Compute Engine service account that Cloud Run runs as needs `roles/aiplatform.user` (call Vertex AI) and `roles/storage.objectUser` (read/write your bucket). The default `Editor` role most new projects start with covers both, so the deploy usually works as-is. If you tightened your project's IAM, grant those two roles explicitly:
> ```bash
> SA=$(gcloud iam service-accounts list \
>  --filter="email:*-compute@developer.gserviceaccount.com" \
>  --format="value(email)")
> gcloud projects add-iam-policy-binding $PROJECT_ID \
>  --member="serviceAccount:$SA" --role="roles/aiplatform.user"
> gcloud projects add-iam-policy-binding $PROJECT_ID \
>  --member="serviceAccount:$SA" --role="roles/storage.objectUser"
> ```


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



