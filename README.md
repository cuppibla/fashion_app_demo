# Fashion App

A Flutter + ADK Go application featuring a virtual fitting room and AI stylist, powered by Gemini and Cloud Run.

---

## Architecture Overview

```
Flutter App  ──── HTTP/REST ────▶  ADK Go Backend (Cloud Run)
                                         │
                              ┌──────────┼──────────┐
                         Fitting Room  Stylist    Catalog
                           Agent       Agent      Agent
                                         │
                              Gemini API + Cloud Storage
```

- **Frontend:** Flutter (web, iOS, Android) in `lib/`
- **Backend:** ADK Go agents in `agents/`; exposes a REST API on port `8080`

---

## Backend Setup (ADK Go Agent Server)

### Prerequisites

- **Go 1.22+** — [install](https://go.dev/doc/install)
- **A Gemini API key** — get one from [Google AI Studio](https://aistudio.google.com/apikey)
- **Google Cloud project** with Cloud Storage enabled
- **A GCS bucket** for artifact storage (generated try-on images)
- **`gcloud` CLI** — [install](https://cloud.google.com/sdk/docs/install), then `gcloud auth login`

### 1. Configure environment variables

```bash
cd agents
cp example.env .env
```

Edit `.env` and fill in all three values:

```env
GEMINI_API_KEY=your_gemini_api_key_here
GOOGLE_CLOUD_PROJECT=your_gcp_project_id
GCS_BUCKET=your_gcs_bucket_name
```

> ⚠️ `.env` is gitignored — never commit it.

### 2. Upload product catalog images to GCS

The agents read product images from your GCS bucket. Run this once (or whenever assets change):

```bash
gcloud storage cp -r ../assets/* gs://$GCS_BUCKET/catalog-assets
```

### 3. Start the backend server

```bash
cd agents
go run . web --write-timeout=300s --read-timeout=60s --idle-timeout=120s api --sse-write-timeout=300s
```

Or use the convenience script:

```bash
cd agents
./run.sh
```

The server starts on **`http://localhost:8080`**. You'll see:

- **REST API** at `http://localhost:8080/api` — used by the Flutter app
- **ADK Dev UI** at `http://localhost:8080` — chat with agents and inspect events/artifacts during development

> **Tip:** The dev UI lets you test agents independently before wiring them to the app. Try chatting with the `fitting room` or `stylist` agent directly!

---

## Frontend Setup (Flutter App)

### Prerequisites

- **Flutter SDK** — [Installation Guide](https://docs.flutter.dev/get-started/install)
- For iOS: [Xcode](https://developer.apple.com/xcode/)
- For Android: [Android Studio](https://developer.android.com/studio)

### 1. Install dependencies

```bash
# From the project root (fashion_app-1/)
flutter pub get
```

### 2. Configure the backend URL

The Flutter app points to the ADK backend via `lib/utils/app_constants.dart`:

```dart
static const String adkBackendUrl = 'http://localhost:8080/api';
```

- **Local development:** leave as-is — matches the backend running on port 8080.
- **Deployed backend:** change this URL to your Cloud Run service URL.

### 3. Run the app

Check available devices:

```bash
flutter devices
```

**Run in Chrome (quickest for local dev):**

```bash
flutter run -d chrome
```

**Run as a web server (useful for testing on other devices on the same network):**

```bash
flutter run -d web-server --web-port=8081
# Open http://localhost:8081 in any browser
```

**Run on iOS Simulator:**

```bash
flutter run -d iPhone   # or use the device ID from `flutter devices`
```

**Run on Android Emulator:**

```bash
flutter run -d emulator-5554   # use the device ID from `flutter devices`
```

> **Hot reload:** While the app is running, press `r` in the terminal to hot-reload changes, or `R` to hot-restart.

### 4. Editing the code?
- **Editor:** Our favorite IDE: [Google Antigravity](https://antigravity.google) with the [Flutter extension](https://marketplace.visualstudio.com/items?itemName=Dart-Code.flutter). We also like [VS Code](https://code.visualstudio.com/) and [Android Studio](https://developer.android.com/studio).

---

## Running Everything Together

Open two terminal tabs:

| Tab | Command | What it does |
|-----|---------|--------------|
| 1 | `cd agents && ./run.sh` | Starts ADK Go backend on `:8080` |
| 2 | `flutter run -d chrome` | Starts Flutter frontend |

Then open your browser — the app will connect to the local backend automatically.

---

## Helpful Resources

- [Flutter Fundamentals](https://docs.flutter.dev/get-started/fundamentals)
- [ADK Go Documentation](https://google.golang.org/adk)
- [Gemini API Docs](https://ai.google.dev/docs)
- [Cloud Run Docs](https://cloud.google.com/run/docs)

---

## 🏗️ AI Workshop Details

Welcome to the **AI Workshop**! In this session, you'll act as an engineer at a fast-growing retail brand. Our existing app is fully developed, and leadership wants to rapidly introduce AI capabilities to dramatically improve the user shopping experience.

You will build out two new highly-requested AI journeys:
1. **Virtual Try-On:** Generate an image of what a specific retail item will look like on the user.
2. **Style Me:** An AI recommender that acts as a personal stylist, generating cohesive outfits.

### Repository Architecture

To make onboarding easy, the workshop isolates the complex legacy code from your workspace. We are operating via a multi-workspace setup.

```text
ai_workshop/
├── adk_backend/                <- Your Go backend workspace
│   ├── agent.go                <- Main agent stubs
│   ├── fittingroom/            <- Virtual Try-On agent logic
│   ├── stylist/                <- Style Me agent logic
│   └── tools/                  <- Agent tools
│
├── flutter_frontend/           <- Your Flutter workspace
│   └── lib/
│       ├── core_app/           <- Our existing retail app—do not modify!
│       └── workshop_tasks/     <- Write your AI flow code here!
│           ├── step_1_try_it_on/
│           └── step_2_style_me/
│
└── solution/                   <- The answer key!
```

### How to approach the codebase
1. **Focus on `flutter_frontend/lib/workshop_tasks/` and `adk_backend/`.** You won't need to touch anything inside `core_app/`.
2. **Follow the breadcrumbs:** The code is heavily annotated. Search your editor for `// TODO: Workshop Step X` to find exactly where you need to write code.
3. **If you get stuck:** Don't panic! Check the `solution/` directory for the completed implementation.

---

### Step-by-Step Guide

#### Step 1: Virtual Try-On Frontend
Start in the frontend! We want a shiny new "Virtual Try On" button on the Product Detail Page.
* **Goal:** Navigate to `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/screens/product_detail_screen.dart` to add the button, then implement the UI flow in `try_it_on_screen.dart`.

#### Step 2: Virtual Try-On Backend
Handle the frontend request by wiring up the Go Agent.
* **Goal:** Navigate to `adk_backend/fittingroom/` and write the ADK instructions handling image uploads and model inference.

#### Step 3: Style Me Frontend
Fashion inspiration mode! The "Style Me" screen loads dynamic, AI-generated outfits based on a specific product.
* **Goal:** Navigate to `flutter_frontend/lib/workshop_tasks/step_2_style_me/ui/screens/style_me_summary_screen.dart` and complete the UI component.

#### Step 4: Style Me Backend Agent
You'll need an agent capable of knowing the catalog and matching garments.
* **Goal:** Navigate to `adk_backend/stylist/` and define the instructions, tool usage, and prompts for the ADK agent.
