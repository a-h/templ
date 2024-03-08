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

Make sure cache is disabled in your browser's dev tools (under the Network tab). 
Browsers want to cache things like css files and it's tricky to debug.

As of v0.2.543, templ in watch mode will generate .txt files, and after the interrupt, templ will delete
those .txt files. You will see templ telling you this after you stop it.

SSG is an additional challenge for air because there are three factors at play: the templ compiler, the tailwindcss
compiler, and the actual go program that wants to spit out html files. Air is responsible for rerunning the SSG program.
It seems like the order of events makes air inconsistent. 
If you're having issues with `npm run dev`, you can still use `npm run start` and restart the process when you're ready to see your changes.

