package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/golang-jwt/jwt/v4"
)

const (
    ServerAddr       = ":8080"
    JWTSecretKey     = "supersecretkey" 
    AuthHeaderPrefix = "Bearer "
)

var secretKey = []byte(JWTSecretKey)

func main() {
    mux := http.NewServeMux()

    mux.Handle("/api/", http.StripPrefix("/api", apiHandler()))

    server := &http.Server{
        Addr:         ServerAddr,
        Handler:      loggingMiddleware(mux),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    serverErrors := make(chan error, 1)

    go func() {
        log.Printf("API Gateway running on %s", ServerAddr)
        serverErrors <- server.ListenAndServe()
    }()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    select {
    case err := <-serverErrors:
        log.Fatalf("Could not start server: %v", err)

    case sig := <-sigChan:
        log.Printf("Received %v signal, initiating graceful shutdown", sig)

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := server.Shutdown(ctx); err != nil {
            log.Fatalf("Graceful shutdown failed: %v", err)
        }

        log.Println("Server gracefully stopped")
    }
}

func apiHandler() http.Handler {
    mux := http.NewServeMux()

    mux.Handle("/auth/", http.StripPrefix("/auth", reverseProxy("http://localhost:8081")))
    mux.Handle("/catalog/", http.StripPrefix("/catalog", reverseProxy("http://localhost:8082")))
    mux.Handle("/cart/", http.StripPrefix("/cart", withJWTAuth(reverseProxy("http://localhost:8083"))))
    mux.Handle("/inventory/", http.StripPrefix("/inventory", withJWTAuth(adminOnly(reverseProxy("http://localhost:8084")))))
    mux.Handle("/ws/join", websocketReverseProxyWithJWT("http://localhost:8085"))
    mux.Handle("/customer/", http.StripPrefix("/customer", withJWTAuth(reverseProxy("http://localhost:8086"))))

    mux.Handle("/", http.NotFoundHandler())

    return mux
}

func reverseProxy(target string) http.Handler {
    targetURL, err := url.Parse(target)
    if err != nil {
        log.Fatalf("Failed to parse target URL %s: %v", target, err)
    }

    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.Header.Set("X-Proxy-By", "Go-API-Gateway")
    }

    proxy.Transport = &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 10 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout: 10 * time.Second,
    }

    return proxy
}

func withJWTAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims, err := validateJWT(r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        if userID, ok := claims["Id"].(string); ok {
            r.Header.Set("X-User-ID", userID)
        }
        if userRole, ok := claims["Role"].(string); ok {
            r.Header.Set("X-User-Role", userRole)
        }
        if userEmail, ok := claims["Email"].(string); ok {
            r.Header.Set("X-User-Email", userEmail)
        }

        next.ServeHTTP(w, r)
    })
}

func adminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userRole := r.Header.Get("X-User-Role")
        if userRole != "admin" {
            http.Error(w, "Access denied: admin only", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func websocketReverseProxyWithJWT(target string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims, err := validateJWT(r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        if userRole, ok := claims["Role"].(string); ok && userRole != "user" && userRole != "admin" {
            http.Error(w, "Access denied: insufficient permissions", http.StatusForbidden)
            return
        }

        if userID, ok := claims["Id"].(string); ok {
            r.Header.Set("X-User-ID", userID)
        }
        if userRole, ok := claims["Role"].(string); ok {
            r.Header.Set("X-User-Role", userRole)
        }
        if userEmail, ok := claims["Email"].(string); ok {
            r.Header.Set("X-User-Email", userEmail)
        }

        targetURL, err := url.Parse(target)
        if err != nil {
            log.Printf("Failed to parse WebSocket target URL %s: %v", target, err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        proxy := httputil.NewSingleHostReverseProxy(targetURL)

        originalDirector := proxy.Director
        proxy.Director = func(req *http.Request) {
            originalDirector(req)
            req.URL.Path = "/ws/join"
        }

        proxy.Transport = &http.Transport{
            DialContext: (&net.Dialer{
                Timeout:   10 * time.Second,
                KeepAlive: 10 * time.Second,
            }).DialContext,
            TLSHandshakeTimeout: 10 * time.Second,
        }

        proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
            log.Printf("WebSocket proxy error: %v", err)
        }

        if isWebSocketRequest(r) {
            proxy.ServeHTTP(w, r)
            return
        }

        http.Error(w, "Not a WebSocket request", http.StatusBadRequest)
    })
}

func validateJWT(r *http.Request) (jwt.MapClaims, error) {
    var tokenString string

    authHeader := r.Header.Get("Authorization")
    if authHeader != "" && strings.HasPrefix(authHeader, AuthHeaderPrefix) {
        tokenString = strings.TrimPrefix(authHeader, AuthHeaderPrefix)
        log.Println("Token found in Authorization header")
    } else {
        tokenString = r.URL.Query().Get("token")
        if tokenString != "" {
            log.Println("Token found in query parameters")
        } else {
            log.Println("Token not found in Authorization header or query parameters")
            return nil, fmt.Errorf("Authorization header or token query parameter is required")
        }
    }

    log.Printf("Token string: %s", tokenString)

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method")
        }
        return secretKey, nil
    })

    if err != nil {
        log.Printf("Token parse error: %v", err)
        return nil, fmt.Errorf("Invalid token: %v", err)
    }

    if !token.Valid {
        log.Println("Token is invalid or expired")
        return nil, fmt.Errorf("Invalid or expired token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        log.Println("Invalid token claims")
        return nil, fmt.Errorf("Invalid token claims")
    }

    if exp, ok := claims["exp"].(float64); ok {
        if int64(exp) < time.Now().Unix() {
            log.Println("Token has expired")
            return nil, fmt.Errorf("Token has expired")
        }
    } else {
        log.Println("Invalid expiration in token")
        return nil, fmt.Errorf("Invalid expiration in token")
    }

    return claims, nil
}

func isWebSocketRequest(r *http.Request) bool {
    connectionHeader := strings.ToLower(r.Header.Get("Connection"))
    upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
    return strings.Contains(connectionHeader, "upgrade") && upgradeHeader == "websocket"
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if isWebSocketRequest(r) {
            next.ServeHTTP(w, r)
            return
        }

        startTime := time.Now()
        lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(lrw, r)
        duration := time.Since(startTime)

        log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, duration)
    })
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}
