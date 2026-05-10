# AI Fashion App Workshop: Try It On & Style Me

## 1. Introduction
Welcome! In this codelab, you will learn how to integrate powerful AI features into an existing Flutter application, chaining two distinct agents together: a **Virtual Try-On** generator and a **Personal Stylist**. 

### What you'll learn
- How to manage asynchronous AI inference states using a `Provider`.
- How to build a state machine router using `AnimatedSwitcher` to switch UI states.
- How to implement an Agentic Handoff, passing multimodal context (images + text) from one AI flow into another.
- How to build an interactive, iterative feedback loop using a chat interface.

### What you'll build
- A Virtual Try-On flow that takes a user photo and a selected clothing item, and visualizes the fit.
- A "Style Me" bottom-sheet brief that collects semantic context (location/occasion).
- A Personal Stylist flow that consumes all previous context to curate outfit carousels, allowing users to converse with the Stylist to refine recommendations.

## 2. Connecting to the ADK Backend (Service Layer)
> 🚧 **Placeholder:** This section will guide the developer through implementing the Dart/Flutter service layers that connect the client app's `TryItOnProvider` and `StylingProvider` to the Go-based ADK backend agent.

## 3. Background: State Management in Flutter
Before we start coding, it is important to understand how Flutter apps are built. **Flutter is declarative**, meaning the application UI is essentially a function of the application's state. When the state changes, the UI rebuilds.

In this workshop, we use **Providers** to manage that state. A Provider acts as the "brain," holding data (like the user's AI-generated image) and functions (like requesting the AI to try on a new shirt). By using `context.watch<OurProvider>()` inside our widgets, Flutter automatically redraws the screen whenever the Provider announces an update.

---

## 4. The Virtual Try-On Experience

### Introduction: Data Models & Services
Before writing UI code, it is important to understand the foundation of the Virtual Try-On flow.

**The `TryOnState` Data Model**
The Try-On experience is governed by a state machine. The `TryOnState` enum defines the stages of the flow:
* `initial`: The user hasn't started.
* `imagePicked` & `generating`: The AI is currently processing the image.
* `success`: The try-on image has been successfully generated.
* `error`: Something went wrong.

**The `TryItOnProvider` & `TryItOnService`**
The provider holds the current `TryOnState`, the user's uploaded image bytes, and the generated result. The `processTryOn` method orchestrates picking a photo and sending the selected product image alongside the user's photo to the backend AI to generate the visualization.

### Step 1: Adding the Virtual Try-On Entry Point
To provide an entry point for users to discover the new "Virtual Try-On" feature, add a button to the **Product Detail Page**.

1. Navigate to `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/1_product_detail_screen.dart`.
2. Locate `// START_WORKSHOP_STEP 1` inside the sticky footer. 
3. Type or copy and paste the following code block to implement the entry point:

```dart
// START_WORKSHOP_STEP 1
// Initializes the Try-On provider with the selected product and navigates to the Try-On flow.
AppIconButton(
  icon: Icons.person_outline,
  onPressed: () {
    context.read<TryItOnProvider>().initializeWithProduct(
      widget.product,
    );
    Navigator.push(
      context,
      FadePageRoute(page: const TryItOnScreen()),
    );
  },
),
const SizedBox(width: 12),
// END_WORKSHOP_STEP
```

**What is this code doing?**
It adds an entry point button next to the "Add to Bag" button. Tapping it sets the chosen product in the provider and routes to our Try-On experience.

> ⚡ **Check Your Work:**
> Type `r` in your terminal to **Hot Reload** the app. Navigate to any product and you should see the new Virtual Try-On icon right next to the "Add to Bag" button! Tap on it to enter your new flow.

---

### 💡 Code Callout 1.5: The Custom App Bar
1. Navigate to `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/3_try_on_app_bar.dart`.
2. Note the `// START_WORKSHOP_CALLOUT 1.5` block. **You don't need to write any code here.**

```dart
// START_WORKSHOP_CALLOUT 1.5
// Only displays the "Change Photo" action if safety conditions are met, and restarts the AI generation pipeline if tapped.
if (provider.userImageBytes != null &&
    provider.state == TryOnState.success)
  ChangePhotoButton(
    onPressed: () async {
      final provider = context.read<TryItOnProvider>();
      final picked = await provider.pickImage();
      if (picked && context.mounted) {
        ...
        final error = await provider.processTryOn(productPath);
        ...
// END_WORKSHOP_CALLOUT
```
**Why does this exist?**
It gives the user an escape hatch. If they don't like the photo they took, tapping this button seamlessly picks a new one and re-triggers the AI generation state machine without losing their context.

---

### Step 2: Routing the Try-On UI States
We need to tell the UI exactly what to display based on the `TryItOnProvider`'s current state.

1. Navigate to `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/2_try_it_on_screen.dart`. 
2. Locate `// START_WORKSHOP_STEP 2`.
3. Type or copy and paste the following `AnimatedSwitcher` to wire up the state routing:

```dart
// START_WORKSHOP_STEP 2
// State Router: Uses pattern matching to display the correct screen based on the AI image generation status.
AnimatedSwitcher(
  duration: AppDurations.medium,
  child: switch (tryOnProvider.state) {
    TryOnState.initial || TryOnState.error => const ChooseImageScreen(),
    TryOnState.imagePicked || TryOnState.generating => LoadingScreen(
      userImage: tryOnProvider.userImageBytes,
    ),
    TryOnState.success => const FittingRoomScreen(),
  },
),
// END_WORKSHOP_STEP 2
```

**What is this code doing?**
It uses Dart 3 pattern matching to turn this screen into a state machine router. As the AI progresses from initializing, to generating, and finally to success, the `AnimatedSwitcher` gracefully fades between the screens.

---

### Step 3: Firing the AI Try-On Request
The user needs a way to upload a picture of themselves. 

1. Navigate to `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/4_user_image_selector.dart`.
2. Locate `// START_WORKSHOP_STEP 3`.
3. Type or copy and paste the following `PrimaryIconButton` to fire the API request:

```dart
// START_WORKSHOP_STEP 3
// The primary action that calls the Provider to start the backend AI Try-On request using the selected product.
PrimaryIconButton(
  onPressed: () async {
    final provider = context.read<TryItOnProvider>();
    final error = await provider.processTryOn(
      provider.products[provider.selectedProductIndex].productImage,
    );
    if (context.mounted) {
      context.showTryOnResult(provider, error);
    }
  },
  icon: const Icon(Icons.photo_library),
  label: 'Upload from Gallery',
),
// END_WORKSHOP_STEP
```

**What is this code doing?**
This triggers the backend inference call. By awaiting `processTryOn`, the UI implicitly enters the loading state (handled in Step 2), and any errors are safely shown using `showTryOnResult`.

---

### Step 4: Displaying the Generated AI Image

1. Navigate to `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/5_fitting_room.dart`.
2. Locate `// START_WORKSHOP_STEP 4`.
3. Type or copy and paste this snippet to subscribe to the provider and fetch the generated image:

```dart
// START_WORKSHOP_STEP 4
// Subscribes to the Provider to fetch and display the generated Try-On image.
final provider = context.watch<TryItOnProvider>();
final generatedImage = provider.generatedImage;
// END_WORKSHOP_STEP 4
```

**What is this code doing?**
It watches the provider so that when the API request finishes and emits the payload, the `Image.memory` widget below (which consumes `generatedImage`) rebuilds automatically to display the Try-On result!

> ⚡ **Check Your Work:**
> Type `r` to **Hot Reload**. Now tap the "Upload from Gallery" button we added in Step 3, select a photo, and watch the UI transition all the way through the state machine until it renders the final try-on composite!

---

### Bonus Step: The Product Selector Carousel
Right beneath Step 4 in `5_fitting_room.dart`:

1. Locate `// START_WORKSHOP_STEP BONUS`.
2. Uncomment (or type out) the `ProductSelectorList` code to enable hot-swapping:

```dart
// START_WORKSHOP_STEP BONUS
ProductSelectorList(
  products: provider.products,
  selectedProductIndex: provider.selectedProductIndex,
  onProductSelected: (index) async {
    if (provider.state != TryOnState.generating) {
      provider.setSelectedProductIndex(index);
      final error = await provider.processTryOn(
        provider.products[index].productImage,
        pickNewImage: false,
      );
      ...
    }
  },
),
// END_WORKSHOP_STEP BONUS
```

**What is this code doing?**
It gives users a horizontal carousel of items to tap. When tapped, it fires off a new generation request instantly without requiring the user to pick their image again.

---

## 5. Agentic Handoff & The Styling Assistant
Now that the user has a Virtual Try-On, we transition them to Agent 2: a Personal Stylist that curates matching outfits for the piece they just bought.

### Introduction: Data Models & Services
To power the personal styling agent, we need robust data models bridging the Try-On session with the conversational UI.

**The `StyleRequest` Data Model**
This model captures the user's semantic intent. It bundles the `location`, `occasion`, and `notes` entered by the user, alongside the `userImageData` (the photo they uploaded in Part 1). Packaging all of this together gives the AI perfect visual and textual context!

**The `Outfit` Data Model**
When the AI responds, it returns a list of `Outfit` objects. Each `Outfit` contains an AI-generated image visualization (`imageData`), a list of the retail `Product` items featured in that image, and a conversational `commentary` string from the stylist explaining why the pieces match.

**The `StylingService`**
The `getStyleSuggestions` method creates the initial `Outfit` list from the `StyleRequest`, while `refineWithFeedback` iterates on those outfits via natural language chat.

---

### Step 5: Invoking the Stylist

1. In `flutter_frontend/lib/workshop_tasks/step_1_try_it_on/ui/5_fitting_room.dart`, scroll down.
2. Locate `// START_WORKSHOP_STEP 5`.
3. Type or copy and paste the `AppOutlinedButton` to connect the user to the styling agent:

```dart
// START_WORKSHOP_STEP 5
// Connects the current screen to the "Style Me" form, allowing the user to request AI styling advice.
AppOutlinedButton(
  onPressed: () => _showStyleMeForm(context),
  icon: const Icon(Icons.auto_awesome, color: Colors.amber),
  label: 'Style Me',
  isExpanded: true,
),
// END_WORKSHOP_STEP
```

**What is this code doing?**
We position a clear "Style Me" action next to "Add to Bag". This button opens a bottom sheet asking for the context of their styling request.

---

### Step 6: The Style Brief Form

1. Navigate to `flutter_frontend/lib/workshop_tasks/step_2_style_me/ui/1_style_me_form_sheet.dart`.
2. Locate `// START_WORKSHOP_STEP 6`.
3. Type or copy and paste the following `AppFormField` inputs:

```dart
// START_WORKSHOP_STEP 6
AppFormField(
  label: 'Location',
  hint: 'e.g., Paris, Beach Resort, Office',
  controller: _locationController,
  icon: Icons.location_on_outlined,
  validator: _validateContext,
  autofocus: true,
),
// ... Continues for Occasion and Notes
// END_WORKSHOP_STEP
```

**What is this code doing?**
We use `AppFormField` inputs to capture semantic context from the user (Where are they going? What is the event?). This metadata heavily anchors the LLM's styling choices.

> ⚡ **Check Your Work:**
> Type `r` to **Hot Reload**. Go generate a try-on image, tap your new "Style Me" button from Step 5, and the bottom-sheet form should gracefully slide up from the bottom of the screen!

---

### 💡 Code Callout 7: The Agent Handoff
This is where the magic happens! **Look closely at these two snippets, you don't need to write code here:**

In `1_style_me_form_sheet.dart`:
```dart
// START_WORKSHOP_CALLOUT 7
Navigator.pop(
  context,
  StyleRequest(
    location: _locationController.text.trim(),
    occasion: _occasionController.text.trim(),
    notes: _notesController.text.trim(),
    userImageData: imageBytes,
  ),
);
// END_WORKSHOP_CALLOUT
```

In `5_fitting_room.dart`:
```dart
// START_WORKSHOP_CALLOUT 7
final result = await showModalBottomSheet<StyleRequest>( ... );

if (result != null && context.mounted) {
  context.read<StylingProvider>().getStyleSuggestions(result);
  Navigator.push(
    context,
    MaterialPageRoute(builder: (context) => const StyleMeSummaryScreen()),
  );
}
// END_WORKSHOP_CALLOUT
```

**Why does this exist?**
This logic performs the **Agentic Handoff**. The first agent (Try-On) collected the generated image. The brief collected location/occasion. `StyleRequest` bundles this multimodal data together and injects it into the *second* agent (`StylingProvider`), ensuring the Stylist LLM knows exactly what the user looks like and where they are going!

---

### Step 8: Visualizing the Brief
1. Navigate to `flutter_frontend/lib/workshop_tasks/step_2_style_me/ui/2_style_me_summary_screen.dart`.
2. Locate `//START_WORKSHOP_STEP 8` in the `ThreadCountAppBar`.
3. Type or copy and paste the action icon:

```dart
//START_WORKSHOP_STEP 8
actionIcon: Icons.info_outline,
onActionPressed: () {
  _showStylingBriefDetails(context);
},
// END_WORKSHOP_STEP
```

**What is this code doing?**
It provides an info icon letting the user review the exact `StyleRequest` context that the AI is currently anchoring its outfit suggestions against.

---

### Step 9: Presenting the Curated UI

1. Still in `2_style_me_summary_screen.dart`.
2. Locate `// START_WORKSHOP_STEP 9`.
3. Type or copy and paste the `AnimatedSwitcher` consumer logic:

```dart
// START_WORKSHOP_STEP 9
body: Consumer<StylingProvider>(
  builder: (context, provider, child) {
    return ScreenWithOverlays(
      body: AnimatedSwitcher(
        // ... (Animation Logic) ...
        child: provider.isLoading
            ? const AppLoadingIndicator( // ... )
            : provider.outfits.isEmpty
            ? AppEmptyState( // ... )
            : provider.outfits.length == 1
            ? OutfitCard(outfit: provider.outfits[0])
            : OutfitCarousel(outfits: provider.outfits),
// END_WORKSHOP_STEP
```

**What is this code doing?**
Much like Step 2, this reacts to the AI inference. It conditionally routes between a loading indicator, an empty state, a single `OutfitCard`, or the swipeable `OutfitCarousel` if the AI generated multiple outfit suggestions!

---

### Step 10: The Feedback Loop

1. At the top of `2_style_me_summary_screen.dart`.
2. Locate `// START_WORKSHOP_STEP 10`.
3. Type or copy and paste the `_handleFeedback` function:

```dart
// START_WORKSHOP_STEP 10
void _handleFeedback(BuildContext context, String text) async {
  final provider = context.read<StylingProvider>();
  await provider.refineWithFeedback(text);
}
// END_WORKSHOP_STEP
```

**What is this code doing?**
An AI feature isn't complete without iterative refinement! This method is triggered by the `FeedbackChatBar` at the bottom of the screen. When the user types *"Make it more casual"*, it sends that feedback directly back into the LLM conversation via the `StylingService` memory buffer, yielding updated `Outfit` cards in real time!

---

## 6. Congratulations

> 🚀 **The Grand Finale!**
> Type `R` (shift+r) in your terminal to **Hot Restart** your entire app. Complete the full flow: select an item, upload a photo to generate a Virtual Try-On, hit "Style Me", submit the form, and finally scroll through the curated outfit carousel. Try giving the Stylist some feedback in the chat bar and watch the AI refine its suggestions!

Congratulations on completing the AI Fashion App Codelab. You have successfully chained multiple agents, piped multimodal data together, and architected beautiful declarative UI based on asynchronous state!
