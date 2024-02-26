templ generate -include-version=false
npx tailwindcss -i ./static/in.css -o ./static/style.css
go run .
