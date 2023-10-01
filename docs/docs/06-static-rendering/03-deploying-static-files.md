# Deploying static files

Once you have built static HTML files with templ, you can serve them on any static site hosting platform, or use a web server to serve them.

Ways you could host your site include:

* Fly.io
* Netlify
* Vercel
* AWS Amplify
* Firebase Hosting

Typically specialist static hosting services are more cost-effective than VM or Docker-based services, due to the less complex compute and networking requirements.

Most require you to commit your code to a source repository, with a build process being triggered on commit, but Fly.io allows you to deploy easily from the CLI.

## fly.io

Fly.io is a provider of hosting that is straightforward to use, and has a generous free tier. Fly.io is Docker-based, so you can easily switch out to a dynamic website if you need to.

Following on from the blog example, all that's required is to add a Dockerfile to the project that copies the contents of the `public` directory into the Docker image, followed by running `flyctl launch` to initialize configuration.

```Dockerfile title="Dockerfile"
FROM pierrezemb/gostatic
COPY ./public/ /srv/http/
ENTRYPOINT ["/goStatic", "-port", "8080"]
```

More detailed documentation is available at https://fly.io/docs/languages-and-frameworks/static/

