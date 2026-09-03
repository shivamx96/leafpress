# Publishing Leafpress sites

Leafpress owns site generation and writes a portable static site to `_site/`
or the configured `build.outputDir`. Publishing is delegated to the hosting
provider's supported CLI or CI integration:

```text
leafpress build
```

This boundary keeps provider authentication, project selection, retries,
deployment readiness, and credential storage in provider-maintained tooling.

## Provider workflows

- GitHub Pages: build in GitHub Actions, upload `_site/` with
  `actions/upload-pages-artifact`, and publish it with `actions/deploy-pages`.
- Netlify: run `netlify deploy --prod --dir=_site` after building.
- Vercel: run `vercel deploy --prod --cwd _site` after building and select the
  target project with Vercel's project-linking or `--project` support.
- Other static hosts: upload the contents of `_site/` and configure the host to
  serve `404.html` for missing routes.

The copyable workflows live in the website deployment guides.

## Canonical URL

Set `site.baseURL` to the final production URL before building. Leafpress uses
it for canonical links, Open Graph URLs, `sitemap.xml`, RSS, and any deployment
subpath such as a GitHub Pages project site:

```json
{
  "site": {
    "baseURL": "https://example.github.io/my-garden"
  }
}
```

## Compatibility and migration

The legacy `deploy` object remains accepted in `leafpress.json` for
configuration compatibility, but Leafpress does not read it. It can be removed
after migrating to provider-native tooling.

Older Leafpress versions stored provider tokens in
`~/.config/leafpress/credentials.json` on macOS and Linux, or the equivalent
roaming application-data directory on Windows. After upgrading, delete that
file if no older Leafpress installation needs it and revoke tokens that are no
longer used. The former `.leafpress-deploy-state.json` file is also unused and
can be removed.
