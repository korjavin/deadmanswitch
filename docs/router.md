# HTTP Router Implementation

This document describes the HTTP routing implementation in the Dead Man's Switch application.

## Overview

The application uses the Gorilla Mux router for HTTP routing. This router provides several advantages over the standard Go HTTP router:

- Path variables (e.g., `/users/{id}`)
- Method-specific routing (GET, POST, etc.)
- Subrouters for grouping routes
- Middleware support

## Router Structure

The router is implemented in the `internal/web/router` package and consists of the following components:

- `Router`: The main router struct that wraps the Gorilla Mux router
- `Route`: A struct that represents a single route in the application
- Helper functions for creating routes (GET, POST, etc.)

## Route Registration

Routes are registered in the `RegisterRoutes` method of the `Router` struct. Routes are grouped by functionality:

- Public routes (no authentication required)
- Protected routes (authentication required)
- API routes

## Middleware

The router supports middleware for authentication and other cross-cutting concerns. The main middleware is:

- `AuthMiddleware`: Checks if the user is authenticated and adds the user to the request context

## Path Variables

Path variables are used to extract parameters from the URL path. For example, `/users/{id}` will match `/users/123` and the `id` variable will be set to `123`.

Path variables can be accessed in handlers using the `mux.Vars` function:

```go
func HandleUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    // ...
}
```

## Example Usage

Here's an example of how to use the router:

```go
// Create a new router
router := router.New(repo)

// Register routes
router.RegisterPublicRoutes(
    indexHandler,
    authHandler,
    passkeyHandler,
    recipientsHandler,
    staticHandler,
)

router.RegisterProtectedRoutes(
    dashboardHandler,
    secretsHandler,
    recipientsHandler,
    apiHandler,
    profileHandler,
    settingsHandler,
    historyHandler,
    twofaHandler,
    passkeyHandler,
    secretQuestionsHandler,
)

// Start the server
server := &http.Server{
    Addr:    ":8080",
    Handler: router.Handler(),
}
server.ListenAndServe()
```

## Testing

The router includes tests for:

- Route registration
- Middleware
- Path variables
- Authentication

Run the tests with:

```bash
go test ./internal/web/router
```

## Migration Guide

If you're migrating from the old router to the new one, here are the main changes:

1. Replace `http.ServeMux` with `router.Router`
2. Replace `HandleFunc` with `GET`, `POST`, etc.
3. Replace string manipulation for path parameters with `mux.Vars`
4. Replace the old authentication middleware with `AuthMiddleware`

## Best Practices

- Group related routes together
- Use descriptive route names
- Use path variables instead of query parameters for resource identifiers
- Use middleware for cross-cutting concerns
- Test your routes
