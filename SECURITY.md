# Security Policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately. Do not open a public issue
or pull request for a security problem.

Preferred channel: GitHub private vulnerability reporting at
<https://github.com/shivamx96/leafpress/security/advisories/new>.

<!-- TODO(maintainer): decide whether to publish an email fallback here.
     If so, consider a dedicated alias rather than a personal address. -->

Include as much of the following as you can:

- the leafpress version (`leafpress version`) or the Core module version;
- the component involved (CLI, `leafpress-render`, install script, website);
- steps to reproduce, ideally with a minimal garden or input;
- the impact you believe it has and any suggested fix.

You will receive an acknowledgement within 3 working days. leafpress is
maintained by one person, so please allow up to 14 days for a triage decision.
Confirmed issues are fixed in the next release and disclosed through a GitHub
security advisory and the changelog. You will be credited unless you prefer
otherwise.

## Supported versions

leafpress is in pre-release. Only the latest release receives security fixes.
Users should upgrade with `leafpress update` or reinstall from
<https://leafpress.in/install.sh>.

| Version | Supported |
|---------|-----------|
| Latest `v1.0.0-rc.N` | Yes |
| Older releases | No |

Once `v1.0.0` ships this table will be updated with a support window.

## Scope

The following are in scope:

- **The `leafpress` CLI.** Path handling in `init`, `new`, `build`, and
  `serve`; the output-directory transaction that replaces `_site/`; symlink
  handling in the content scanner.
- **The `leafpress-render` bridge and the Core library.** Input validation
  for host-supplied slugs, configuration, and Markdown, as described in
  `docs/05_RENDERER_CONTRACT.md`.
- **Self-update.** `leafpress update` downloads release archives and verifies
  them against the published `checksums.txt`. Anything that lets a tampered
  archive pass verification is in scope.
- **The install script** at `website/static/install.sh`, served from
  <https://leafpress.in/install.sh>.
- **Generated sites.** Cross-site scripting or content injection that arises
  from leafpress's own templates, client scripts, search index, or graph data
  rather than from raw HTML the author wrote.
- **Release integrity.** Anything affecting the tag verification, binary
  scanning, or checksum publication in `.github/workflows/release.yml`.

## Out of scope

These are design decisions or outside leafpress's control. Reports are still
welcome if you think a default should change, but they are not treated as
vulnerabilities.

- **Raw HTML in Markdown is rendered as written.** leafpress is a
  single-author publishing tool. Authors are trusted, so `<script>` and other
  HTML in a note reaches the output. Hosts that render untrusted Markdown
  through `leafpress-render` are responsible for sanitising input or output.
- **`leafpress serve` exposed on a non-loopback interface.** The preview
  server binds to `127.0.0.1` by default. Exposing it deliberately with
  another host is the operator's choice; the server is a development tool,
  not a production web server.
- **Third-party hosting.** Behaviour of GitHub Pages, Netlify, Vercel, or any
  host serving the generated `_site/` directory.
- **Vulnerabilities in dependencies with no reachable path.** CI runs
  `govulncheck` on source and on every release binary. Findings that the
  scanner shows as unreachable are tracked through Dependabot rather than as
  security reports.
- **Denial of service by pathological content** in the author's own garden,
  such as extremely large or deeply nested notes.

## How leafpress protects users

- Release binaries are built from tagged commits only after the full test
  suite passes and both Go modules and every platform binary pass
  `govulncheck`.
- The release publishes a single `checksums.txt`. The install script and the
  self-updater refuse to install an archive whose SHA-256 does not match.
- Both Go modules are scanned in CI on every push, and Dependabot proposes
  dependency and GitHub Actions updates weekly.
- The build rejects routes that are not portable, symlinks that escape the
  garden, and output directories it does not own before writing anything.

## Disclosure

Fixed vulnerabilities are announced in `website/changelog.md`, in the GitHub
release notes, and where applicable as a GitHub security advisory with a CVE
requested through GitHub. Reports against Go dependencies reference the Go
vulnerability database identifier, for example GO-2026-5970.
