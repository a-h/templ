# Templ docs generator

To edit docs, edit the relevant markdown files in `docs/`.

To see your changes locally, run `npm run install`, then `npm run start`. 
You will need go, npm (installed from node), and [xc](https://xcfile.dev/getting-started/#installation) installed.

This will not watch for file changes, so when you make changes you will need to stop the server and run `npm run build` again.

To build the docs for production, `cd` into this directory, run `npm install`, then `npm run build`.

### Prism

The Prism we install uses the following languages
- Markup + HTML + XML + SVG + MathML + SSML + Atom + RSS 2.78KB
- CSS 1.2KB
- C-like 0.69KB
- JavaScript 4.5KB
- Bash + Shell + Shell zeitgeist 6KB
- Go arnehormann 0.95KB
- Go module RunDevelopment 0.41KB
- React JSX vkbansal 2.33KB
- React TSX 0.3KB
- JSON + Web App Manifest CupOfTea 0.44KB
- Lua Golmote 0.58KB
- Docker JustinBeckwith 1.49KB
- TypeScript vkbansal 1.26KB
- YAML hason 1.92KB

JS files downloaded from these links
- https://prismjs.com/download.html#themes=prism-solarizedlight&languages=markup+css+clike+javascript+bash+docker+go+go-module+json+lua+nix+jsx+tsx+typescript+yaml
- https://prismjs.com/download.html#themes=prism-tomorrow&languages=markup+css+clike+javascript+bash+docker+go+go-module+json+lua+nix+jsx+tsx+typescript+yaml

CSS files downloaded from @kenchandev's fantastic Prism theme here: 
- https://github.com/kenchandev/prism-theme-one-light-dark
