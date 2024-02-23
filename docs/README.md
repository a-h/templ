# # Templ docs generator

To edit docs, edit the relevant markdown files in `docs/`.

To build the docs for production, `cd` into this directory, run `npm install` then `npm run build`.

To run the docs in localhost, run `npm run dev`. This will not watch for file changes, 
>so when you make changes you will need to stop the server and run `npm run build` again

`npm run build` is a shortcut for the following three commands:
```sh
tailwindcss -i ./static/in.css -o ./static/style.css
templ generate ./src/components
go run main.go
```

`npm run start` is a shortcut for the following three commands:
```sh
tailwindcss -i ./static/in.css -o ./static/style.css
templ generate ./src/components
go run main.go --local
```