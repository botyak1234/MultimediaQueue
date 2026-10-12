package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"media-planner/backend/auth"
	"media-planner/backend/db"
	"media-planner/backend/media"
	"media-planner/backend/planner"
	"media-planner/backend/queue"
)

//go:embed frontend/dist/*
var embeddedFiles embed.FS

func main() {
	database, err := db.InitDB("data.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	authHandler := &auth.AuthHandler{DB: database}
	mediaHandler := &media.MediaHandler{DB: database}
	queueHandler := &queue.QueueHandler{DB: database}
	plannerHandler := &planner.PlannerHandler{DB: database}

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/auth/me", authHandler.Middleware(authHandler.Me))

	// Media routes
	mux.HandleFunc("GET /api/media", authHandler.Middleware(mediaHandler.ListMedia))
	mux.HandleFunc("POST /api/media", authHandler.Middleware(mediaHandler.CreateMedia))
	mux.HandleFunc("DELETE /api/media/{id}", authHandler.Middleware(mediaHandler.DeleteMedia))
	mux.HandleFunc("PUT /api/media/{id}/progress", authHandler.Middleware(mediaHandler.UpdateProgress))

	// Queue routes
	mux.HandleFunc("GET /api/queue", authHandler.Middleware(queueHandler.ListQueue))
	mux.HandleFunc("POST /api/queue", authHandler.Middleware(queueHandler.AddToQueue))
	mux.HandleFunc("PUT /api/queue/{id}", authHandler.Middleware(queueHandler.UpdateQueueStatus))
	mux.HandleFunc("DELETE /api/queue/{id}", authHandler.Middleware(queueHandler.DeleteFromQueue))

	// Planner route
	mux.HandleFunc("POST /api/planner", authHandler.Middleware(plannerHandler.GeneratePlan))

	// Static SPA Serving
	distFS, err := fs.Sub(embeddedFiles, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to sub embedded files: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to open file in embedded dist
		f, err := distFS.Open(r.URL.Path[1:])
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("\n======================================================\n")
	fmt.Printf("🚀 Media Planner успешно запущен!\n")
	fmt.Printf("🔗 Откройте в браузере: %s\n", url)
	fmt.Printf("======================================================\n\n")

	// Open browser automatically
	openBrowser(url)

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Printf("Не удалось автоматически открыть браузер: %v", err)
	}
}
