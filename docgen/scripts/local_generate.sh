templ generate -include-version=false
npx -p tailwindcss -p @tailwindcss/typography tailwindcss -i ./static/in.css -o ./static/style.css
go run . --url http://localhost:8080/