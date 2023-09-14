# Hosting using Docker

Applications that use templ can be deployed using the same techniques and platforms as any other Go application.

An example Dockerfile is provided in the https://github.com/a-h/templ/tree/main/examples/counter-basic example.

# Static content

### Adding static content to the Docker container

Web applications often need to include static content such as CSS, images, and icon files.

The https://github.com/a-h/templ/tree/main/examples/counter-basic example has an `assets` directory for this purpose.

The `COPY` instruction in the Dockerfile copies all of the code and the `assets` directory to the container so that it can be served by the application.

```Dockerfile title="Dockerfile"
# Build.
FROM golang:1.20 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
// highlight-next-line
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint

# Deploy.
FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /
COPY --from=build-stage /entrypoint /entrypoint
// highlight-next-line
COPY --from=build-stage /app/assets /assets
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/entrypoint"]
```

### Serving static content

Once the `/assets` directory has been added to the deployment Docker container, the `http.FileServer` function must be used to serve the content.

```go title="main.go"
func main() {
	// Initialize the session.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()

	// Handle POST and GET requests.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postHandler(w, r)
			return
		}
		getHandler(w, r)
	})

	// Include the static content.
	// highlight-next-line
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Add the middleware.
	muxWithSessionMiddleware := sessionManager.LoadAndSave(mux)

	// Start the server.
	fmt.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", muxWithSessionMiddleware); err != nil {
		log.Printf("error listening: %v", err)
	}
}
```

## Example deployment

The https://github.com/a-h/templ/tree/main/examples/counter-basic example is deployed at https://counter-basic.fly.dev/

:::note
This sample application stores the counts in RAM. If the server restarts, all of the information is lost. To avoid this, use a data store such as DynamoDB or Cloud Firestore. See https://github.com/a-h/templ/tree/main/examples/counter for an example of this.
:::

