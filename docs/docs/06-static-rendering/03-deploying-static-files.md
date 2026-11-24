# Deploying static files

Once you have built static HTML files with templ, you can serve them on any static site hosting platform, or use a web server to serve them.

Ways you could host your site include:

* Fly.io
* Netlify
* Vercel
* AWS Amplify
* Firebase Hosting
* deed.page

Typically specialist static hosting services are more cost-effective than VM or Docker-based services, due to the less complex compute and networking requirements.

Most require you to commit your code to a source repository, with a build process being triggered on commit, but Fly.io and deed.page allow you to deploy easily from the CLI.

## fly.io

Fly.io is a provider of hosting that is straightforward to use, and has a generous free tier. Fly.io is Docker-based, so you can easily switch out to a dynamic website if you need to.

Following on from the blog example, all that's required is to add a Dockerfile to the project that copies the contents of the `public` directory into the Docker image, followed by running `flyctl launch` to initialize configuration.

```Dockerfile title="Dockerfile"
FROM pierrezemb/gostatic
COPY ./public/ /srv/http/
ENTRYPOINT ["/goStatic", "-port", "8080"]
```

More detailed documentation is available at https://fly.io/docs/languages-and-frameworks/static/

## deed.page

[deed.page](https://deed.page) hosts static files uploaded with a single HTTP request. It doesn't need an account, a source repository or a Docker image.

Following on from the blog example, pack the contents of the `public` directory into a tarball and upload it with `curl`. Choose your own slug, because it becomes the site's address.

```bash
tar -C public -czf site.tar.gz .
curl -T site.tar.gz "https://deed.page/v1/sites?slug=my-templ-blog"
```

The JSON response contains the site URL (`https://my-templ-blog.deed.page/`) and a token. Keep the token: to publish an update to the same URL, rebuild the site and run the same `curl` command with `-H "Authorization: Bearer $DEEDPAGE_TOKEN"` added.

A single upload is limited to 4 MB. Larger sites and other options are covered in the [deed.page documentation](https://deed.page/docs).
