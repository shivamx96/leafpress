package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/shivamx96/leafpress/cli/internal/build"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
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
	rebuildMu sync.Mutex

	// File watcher
	watcher *fsnotify.Watcher

	// ignore mirrors the scanner's exclusion rules so watch mode and a full
	// build agree on what counts as content.
	ignore *content.IgnoreMatcher
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
	// Compile the ignore globs once; Config.Validate has already reported a
	// malformed pattern, so this is belt and braces.
	ignore, err := content.NewIgnoreMatcher(s.cfg.Build.Ignore)
	if err != nil {
		return err
	}
	s.ignore = ignore

	// Find available port
	port := s.cfg.Build.Port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Try to find another port
		for i := 1; i <= 10; i++ {
			port = s.cfg.Build.Port + i
			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				fmt.Printf("Port %d in use, using %d instead\n", s.cfg.Build.Port, port)
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
	outputDir := filepath.Join(cwd, s.cfg.Build.OutputDir)
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
	fmt.Println("  Press Ctrl+C to stop")

	return server.Serve(listener)
}

// handleStatic serves static files with live reload script injection
func (s *Server) handleStatic(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Disable caching for development
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Clean and sanitize path to prevent directory traversal
		urlPath := filepath.Clean(r.URL.Path)
		// Remove leading slash and any .. components
		urlPath = strings.TrimPrefix(urlPath, "/")
		// filepath.Clean on relative path removes .. that would escape
		urlPath = filepath.Clean(urlPath)
		// Reject any path that still tries to escape
		if strings.HasPrefix(urlPath, "..") {
			http.NotFound(w, r)
			return
		}

		// Handle index files
		if urlPath == "" || urlPath == "." {
			urlPath = "index.html"
		} else if !strings.Contains(filepath.Base(urlPath), ".") {
			// Clean URL - try adding /index.html
			urlPath = filepath.Join(urlPath, "index.html")
		}

		filePath := filepath.Join(root, urlPath)

		// Final safety check: ensure resolved path is within root
		absRoot, _ := filepath.Abs(root)
		absFile, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFile, absRoot+string(filepath.Separator)) && absFile != absRoot {
			http.NotFound(w, r)
			return
		}

		// Check if file exists
		info, err := os.Stat(filePath)
		if err != nil {
			// Try with .html extension
			if !strings.HasSuffix(urlPath, ".html") {
				htmlPath := filePath + ".html"
				if _, err := os.Stat(htmlPath); err == nil {
					filePath = htmlPath
					info, _ = os.Stat(filePath)
				}
			}
		}

		if err != nil || info.IsDir() {
			// Serve custom 404.html if it exists
			notFoundPath := filepath.Join(root, "404.html")
			if content, readErr := os.ReadFile(notFoundPath); readErr == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusNotFound)
				w.Write(s.injectLiveReload(content))
				return
			}
			http.NotFound(w, r)
			return
		}

		// For HTML files, read and inject live reload script
		if strings.HasSuffix(filePath, ".html") {
			content, err := os.ReadFile(filePath)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(s.injectLiveReload(content))
			return
		}

		// For all other files, use http.ServeFile (handles content type,
		// range requests for media, caching headers, etc.)
		http.ServeFile(w, r, filePath)
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
	clientCount := len(s.clients)
	s.clientsMu.Unlock()

	if s.opts.Verbose {
		log.Printf("Live reload: browser connected (%d total)", clientCount)
	}

	// Keep connection open
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			s.clientsMu.Lock()
			delete(s.clients, conn)
			clientCount = len(s.clients)
			s.clientsMu.Unlock()
			conn.Close()
			if s.opts.Verbose {
				log.Printf("Live reload: browser disconnected (%d remaining)", clientCount)
			}
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
	// Keep every path observed during the debounce window. The watcher loop
	// performs rebuilds serially, and rebuildMu also protects against rebuilds
	// initiated elsewhere.
	var timer *time.Timer
	var timerC <-chan time.Time
	pending := make(map[string]build.ChangeType)

	// Get working directory for relative path calculation
	cwd, _ := os.Getwd()

	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				if timer != nil {
					timer.Stop()
				}
				return
			}

			// Determine change type
			var changeType build.ChangeType
			if event.Op&fsnotify.Remove != 0 {
				changeType = build.ChangeDelete
			} else if event.Op&fsnotify.Create != 0 {
				changeType = build.ChangeCreate
			} else if event.Op&fsnotify.Write != 0 {
				changeType = build.ChangeModify
			} else {
				continue
			}

			if changeType == build.ChangeCreate {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := s.addWatchDirs(event.Name); err != nil && s.opts.Verbose {
						log.Printf("Failed to watch new directory %s: %v", event.Name, err)
					}
					continue
				}
			}

			// Get relative path for checking
			relPath, err := filepath.Rel(cwd, event.Name)
			if err != nil {
				relPath = event.Name
			}

			// Check if it's a file we care about
			ext := filepath.Ext(event.Name)
			base := filepath.Base(event.Name)
			isStaticFile := isStaticTree(relPath)
			if ext != ".md" && ext != ".css" && base != "leafpress.json" && !isStaticFile {
				continue
			}

			// Reserved names and ignore globs are not content. Without this
			// an edit under docs/ or drafts/ published a page that the next
			// full build would drop.
			if !isStaticFile && base != "leafpress.json" &&
				content.IsExcluded(relPath, s.ignore) {
				if s.opts.Verbose {
					log.Printf("Ignoring change in excluded path: %s", relPath)
				}
				continue
			}

			if s.opts.Verbose {
				log.Printf("File changed: %s (type: %d)", relPath, changeType)
			}

			pending[event.Name] = mergeChangeType(pending[event.Name], changeType)
			if timer == nil {
				timer = time.NewTimer(100 * time.Millisecond)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(100 * time.Millisecond)
			}
			timerC = timer.C

		case <-timerC:
			paths := make([]string, 0, len(pending))
			for path := range pending {
				paths = append(paths, path)
			}
			sort.Strings(paths)
			for _, path := range paths {
				s.rebuildIncremental(path, pending[path])
			}
			clear(pending)
			timerC = nil

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func mergeChangeType(previous, next build.ChangeType) build.ChangeType {
	if next == build.ChangeDelete || next == build.ChangeCreate {
		return next
	}
	if previous == build.ChangeCreate || previous == build.ChangeDelete {
		return previous
	}
	return build.ChangeModify
}

// rebuild rebuilds the site and notifies clients
func (s *Server) rebuild() {
	s.rebuildMu.Lock()
	defer s.rebuildMu.Unlock()

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

func (s *Server) rebuildIncremental(changedPath string, changeType build.ChangeType) {
	s.rebuildMu.Lock()
	defer s.rebuildMu.Unlock()

	// Get relative path for display
	cwd, _ := os.Getwd()
	relPath, _ := filepath.Rel(cwd, changedPath)
	if relPath == "" {
		relPath = changedPath
	}

	if s.opts.Verbose {
		fmt.Printf("Rebuilding (%s)...\n", relPath)
	} else {
		fmt.Println("Rebuilding...")
	}
	start := time.Now()

	stats, err := s.builder.RebuildIncremental(changedPath, changeType)
	if err != nil {
		fmt.Printf("Build error: %v\n", err)
		return
	}

	elapsed := time.Since(start)
	if stats.FullRebuild {
		fmt.Printf("Full rebuild in %s\n", elapsed.Round(time.Millisecond))
	} else {
		fmt.Printf("Rebuilt %d pages in %s\n", stats.PagesRebuilt, elapsed.Round(time.Millisecond))
	}

	// Notify connected browsers to reload
	s.clientsMu.Lock()
	clientCount := len(s.clients)
	s.clientsMu.Unlock()

	if clientCount > 0 {
		s.notifyClients()
		if s.opts.Verbose {
			fmt.Printf("Notified %d browser(s) to reload\n", clientCount)
		}
	}
}

// isStaticTree reports whether a project-relative path is static/ or inside
// it. That tree is reserved for the scanner but still watched: its files are
// copied verbatim into the site.
func isStaticTree(relPath string) bool {
	return relPath == "static" ||
		strings.HasPrefix(relPath, "static"+string(filepath.Separator))
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

		// Prune exactly what the content scan prunes. static/ is the one
		// reserved tree still worth watching: its files are copied into the
		// site, so edits there must trigger a rebuild.
		rel, relErr := filepath.Rel(root, path)
		if relErr == nil && !isStaticTree(rel) && content.IsExcluded(rel, s.ignore) {
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
