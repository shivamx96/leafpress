package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/shivamx96/leafpress/internal/build"
	"github.com/shivamx96/leafpress/internal/config"
)

// Options configures the server
type Options struct {
	Verbose bool
}

// Server handles the development server with live reload
type Server struct {
	cfg     *config.Config
	builder *build.Builder
	opts    Options

	// WebSocket connections for live reload
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex

	// File watcher
	watcher *fsnotify.Watcher
}

// New creates a new development server
func New(cfg *config.Config, builder *build.Builder, opts Options) *Server {
	return &Server{
		cfg:     cfg,
		builder: builder,
		opts:    opts,
		clients: make(map[*websocket.Conn]bool),
	}
}

// Start starts the development server
func (s *Server) Start() error {
	// Find available port
	port := s.cfg.Port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Try to find another port
		for i := 1; i <= 10; i++ {
			port = s.cfg.Port + i
			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				fmt.Printf("Port %d in use, using %d instead\n", s.cfg.Port, port)
				break
			}
		}
		if err != nil {
			return fmt.Errorf("could not find available port: %w", err)
		}
	}

	// Set up file watcher
	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer s.watcher.Close()

	// Watch for file changes
	go s.watchFiles()

	// Add directories to watch
	cwd, _ := os.Getwd()
	if err := s.addWatchDirs(cwd); err != nil {
		return fmt.Errorf("failed to set up file watching: %w", err)
	}

	// Set up HTTP handlers
	mux := http.NewServeMux()

	// Live reload WebSocket endpoint
	mux.HandleFunc("/_lr", s.handleWebSocket)

	// Serve static files with live reload injection
	outputDir := filepath.Join(cwd, s.cfg.OutputDir)
	mux.HandleFunc("/", s.handleStatic(outputDir))

	server := &http.Server{
		Handler: mux,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down...")
		server.Close()
	}()

	fmt.Printf("\n  Server running at http://localhost:%d\n", port)
	fmt.Println("  Press Ctrl+C to stop\n")

	return server.Serve(listener)
}

// handleStatic serves static files with live reload script injection
func (s *Server) handleStatic(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Clean the path
		if path == "/" {
			path = "/index.html"
		} else if !strings.Contains(filepath.Base(path), ".") {
			// Clean URL - try adding /index.html
			path = filepath.Join(path, "index.html")
		}

		filePath := filepath.Join(root, path)

		// Check if file exists
		info, err := os.Stat(filePath)
		if err != nil {
			// Try with .html extension
			if !strings.HasSuffix(path, ".html") {
				htmlPath := filePath + ".html"
				if _, err := os.Stat(htmlPath); err == nil {
					filePath = htmlPath
					info, _ = os.Stat(filePath)
				}
			}
		}

		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		// Read file
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set content type
		contentType := "application/octet-stream"
		switch filepath.Ext(filePath) {
		case ".html":
			contentType = "text/html; charset=utf-8"
		case ".css":
			contentType = "text/css; charset=utf-8"
		case ".js":
			contentType = "application/javascript"
		case ".json":
			contentType = "application/json"
		case ".png":
			contentType = "image/png"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".gif":
			contentType = "image/gif"
		case ".svg":
			contentType = "image/svg+xml"
		case ".ico":
			contentType = "image/x-icon"
		}
		w.Header().Set("Content-Type", contentType)

		// Inject live reload script for HTML files
		if strings.HasSuffix(filePath, ".html") {
			content = s.injectLiveReload(content)
		}

		w.Write(content)
	}
}

// injectLiveReload injects the live reload script before </body>
func (s *Server) injectLiveReload(content []byte) []byte {
	script := `<script>
(function() {
  var ws = new WebSocket('ws://' + location.host + '/_lr');
  ws.onmessage = function() { location.reload(); };
  ws.onclose = function() {
    console.log('Live reload disconnected. Retrying...');
    setTimeout(function() { location.reload(); }, 1000);
  };
})();
</script>`

	html := string(content)
	if idx := strings.LastIndex(html, "</body>"); idx != -1 {
		html = html[:idx] + script + "\n" + html[idx:]
	} else {
		html += script
	}
	return []byte(html)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// handleWebSocket handles WebSocket connections for live reload
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	// Keep connection open
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			s.clientsMu.Lock()
			delete(s.clients, conn)
			s.clientsMu.Unlock()
			conn.Close()
			break
		}
	}
}

// notifyClients sends reload signal to all connected clients
func (s *Server) notifyClients() {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// watchFiles watches for file changes and triggers rebuilds
func (s *Server) watchFiles() {
	// Debounce timer
	var timer *time.Timer
	var mu sync.Mutex

	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			// Only care about write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Check if it's a file we care about
			ext := filepath.Ext(event.Name)
			if ext != ".md" && ext != ".css" {
				continue
			}

			// Debounce
			mu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(100*time.Millisecond, func() {
				s.rebuild()
			})
			mu.Unlock()

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// rebuild rebuilds the site and notifies clients
func (s *Server) rebuild() {
	fmt.Println("Rebuilding...")
	start := time.Now()

	stats, err := s.builder.Build()
	if err != nil {
		fmt.Printf("Build error: %v\n", err)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("Built %d pages in %s\n", stats.PageCount, elapsed.Round(time.Millisecond))

	s.notifyClients()
}

// addWatchDirs recursively adds directories to the watcher
func (s *Server) addWatchDirs(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip non-directories
		if !info.IsDir() {
			return nil
		}

		// Skip output and hidden directories
		name := info.Name()
		if name == "_site" || name == ".leafpress" || name == ".git" ||
			name == "node_modules" || name == ".obsidian" {
			return filepath.SkipDir
		}

		// Add to watcher
		if err := s.watcher.Add(path); err != nil {
			if s.opts.Verbose {
				log.Printf("Failed to watch %s: %v", path, err)
			}
		}

		return nil
	})
}
