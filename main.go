package main

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "strings"
    "time"
)

type Route struct {
    Path   string
    Target string
}

type Gateway struct {
    routes []Route
}

func NewGateway() *Gateway {
    return &Gateway{
        routes: []Route{},
    }
}

func (g *Gateway) AddRoute(path, target string) {
    g.routes = append(g.routes, Route{
        Path:   path,
        Target: target,
    })
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Add CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    
    // Log request
    log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
    
    // Find matching route
    for _, route := range g.routes {
        if strings.HasPrefix(r.URL.Path, route.Path) {
            g.proxyRequest(w, r, route)
            return
        }
    }
    
    // No route found
    if r.URL.Path == "/" || r.URL.Path == "/health" {
        g.handleRoot(w, r)
        return
    }
    
    http.NotFound(w, r)
}

func (g *Gateway) proxyRequest(w http.ResponseWriter, r *http.Request, route Route) {
    target, err := url.Parse(route.Target)
    if err != nil {
        http.Error(w, "Invalid target URL", http.StatusInternalServerError)
        return
    }
    
    // Create reverse proxy
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    // Modify request
    r.URL.Host = target.Host
    r.URL.Scheme = target.Scheme
    r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
    r.Host = target.Host
    
    // Strip route prefix from path
    r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Path)
    if !strings.HasPrefix(r.URL.Path, "/") {
        r.URL.Path = "/" + r.URL.Path
    }
    
    // Serve the request
    proxy.ServeHTTP(w, r)
}

func (g *Gateway) handleRoot(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/health" {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"status":"healthy","service":"api-gateway","timestamp":"%s"}`, 
            time.Now().Format(time.RFC3339))
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    response := `{
    "service": "API Gateway",
    "version": "1.0.0",
    "routes": [`
    
    for i, route := range g.routes {
        if i > 0 {
            response += ","
        }
        response += fmt.Sprintf(`
        {
            "path": "%s",
            "target": "%s"
        }`, route.Path, route.Target)
    }
    
    response += `
    ]
}`
    
    fmt.Fprint(w, response)
}

func main() {
    gateway := NewGateway()
    
    // Configure routes from environment or defaults
    gateway.AddRoute("/api/users", getEnv("USERS_SERVICE", "http://users-service:8080"))
    gateway.AddRoute("/api/posts", getEnv("POSTS_SERVICE", "http://posts-service:8080"))
    gateway.AddRoute("/api/auth", getEnv("AUTH_SERVICE", "http://auth-service:8080"))
    
    port := getEnv("PORT", "8080")
    
    fmt.Printf("API Gateway starting on port %s\n", port)
    fmt.Println("Routes configured:")
    for _, route := range gateway.routes {
        fmt.Printf("  %s -> %s\n", route.Path, route.Target)
    }
    
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      gateway,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
    
    log.Fatal(server.ListenAndServe())
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}