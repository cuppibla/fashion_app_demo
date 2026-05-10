package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"goagents/catalog"
	"goagents/fittingroom"
	"goagents/rootagent"
	"goagents/stylist"
	"goagents/tools"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/artifact/gcsartifact"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/web/webui"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"
)

var port = flag.Int("port", 8080, "local port to listen on.")
var enableWebUI = flag.Bool("webui", false, "start the ADK WebUI with the API server")

func main() {
	flag.Parse()
	godotenv.Load()
	apikey := os.Getenv("GEMINI_API_KEY")
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	bucket := os.Getenv("GCS_BUCKET")

	if apikey == "" {
		log.Fatal("GEMINI_API_KEY must be set.")
	}
	if project == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT must be set.")
	}
	if bucket == "" {
		log.Fatal("GCS_BUCKET must be set.")
	}

	artifacts, err := gcsartifact.NewService(context.Background(),
		bucket)
	// TODO: retry with api key after key restrictions propagage.
	// bucket, option.WithAPIKey(apikey))
	if err != nil {
		log.Fatalf("Failed to create artifact storage: %v", err)
	}
	// artifacts := artifact.InMemoryService()

	catagent, err := catalog.NewCatalogAgent(apikey, "catalog/catalog.yaml")
	if err != nil {
		log.Fatalf("Failed to create catalog agent: %v", err)
	}

	// load agents
	fitagent, err := fittingroom.NewFittingRoomAgent(apikey, catagent)
	if err != nil {
		log.Fatalf("Failed to create fitting agent: %v", err)
	}

	stylistAgent, err := stylist.NewStylistAgent(apikey, catagent)
	if err != nil {
		log.Fatalf("Failed to create stylist agent: %v", err)
	}

	ragent, err := rootagent.NewRootAgent(apikey, fitagent, catagent, stylistAgent)
	if err != nil {
		log.Fatalf("Failed to create root agent: %v", err)
	}

	loader, err := agent.NewMultiLoader(ragent, fitagent, catagent, stylistAgent)
	if err != nil {
		log.Fatalf("Failed to create agent loader: %v", err)
	}

	restHandler, err := adkrest.NewServer(adkrest.ServerConfig{
		SessionService:  session.InMemoryService(),
		MemoryService:   memory.InMemoryService(),
		AgentLoader:     loader,
		ArtifactService: artifacts,
		SSEWriteTimeout: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to make rest server: %v", err)
	}

	r := mux.NewRouter()

	// setup the webui, if we want it.
	if *enableWebUI {
		wui := webui.NewLauncher()
		_, err = wui.Parse([]string{
			"--api_server_address", fmt.Sprintf("http://localhost:%d/api", *port),
		})
		if err != nil {
			log.Fatalf("Failed to initialize webui from flags : %v", err)
		}
		err = wui.SetupSubrouters(r, &launcher.Config{})
		if err != nil {
			log.Fatalf("Failed to initialize webui: %v", err)
		}
	}
	r.Use(tools.LocalhostCORS)
	r.PathPrefix("/api/").Handler(
		http.StripPrefix("/api", tools.LogHandler(restHandler)))
	s := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}
	log.Fatal(s.ListenAndServe())
}
