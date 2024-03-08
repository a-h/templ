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

`npm run dev` is a shortcut for running `dev.sh` script. It runs [air](https://github.com/cosmtrek/air),
templ watch, and tailwind watch. After using ^C, it stops all of them. Air will rebuild the site and restart the 
http server after making a change in a file. To see this change, you will still need to refresh in your browser.

If you notice that, for example, you change a tailwind class, air restarts the server, and you don't see a change, this
is relatively normal. It's not clear what's to blame here, but make sure cache is disabled in your browser's
dev tools (under the Network tab). Browsers want to cache things like css files and it's tricky to debug.

If air doesn't play right for you, you can still use `npm run start`, stop it, and start it again between changes.

As of v0.2.543, templ in watch mode will generate .txt files, and after the interrupt, templ will delete
those .txt files. You will see templ telling you this after you stop it.
