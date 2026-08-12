import { spawnSync } from "node:child_process";
import {
  cpSync,
  existsSync,
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync
} from "node:fs";
import { createServer } from "node:http";
import { tmpdir } from "node:os";
import { dirname, extname, join, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const host = "127.0.0.1";
const port = Number.parseInt(process.env.LEAFPRESS_CONFORMANCE_PORT ?? "4173", 10);
const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const fixtureRoot = join(repoRoot, "testdata", "theme-garden");
const workRoot = mkdtempSync(join(tmpdir(), "leafpress-theme-conformance-"));
const publicRoot = join(workRoot, "public");
const binaryPath = join(workRoot, "leafpress");

const themes = ["classic", "aurora"];
const navStyles = ["base", "sticky", "glassy"];
const activeStyles = ["base", "underlined", "box"];

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: repoRoot,
    encoding: "utf8",
    ...options
  });
  if (result.status !== 0) {
    process.stderr.write(result.stdout ?? "");
    process.stderr.write(result.stderr ?? "");
    throw new Error(`${command} ${args.join(" ")} exited with status ${result.status}`);
  }
}

function copyFixture(destination) {
  cpSync(fixtureRoot, destination, {
    recursive: true,
    filter(source) {
      return relative(fixtureRoot, source).split(sep)[0] !== "_site";
    }
  });
}

function buildFixtures() {
  mkdirSync(publicRoot, { recursive: true });
  run("go", ["build", "-o", binaryPath, "./cli/cmd/leafpress"]);

  for (const theme of themes) {
    for (const navStyle of navStyles) {
      for (const activeStyle of activeStyles) {
        const fixtureName = `${theme}-${navStyle}-${activeStyle}`;
        const sourceRoot = join(workRoot, "sources", fixtureName);
        copyFixture(sourceRoot);

        const configPath = join(sourceRoot, "leafpress.json");
        const config = JSON.parse(readFileSync(configPath, "utf8"));
        config.site.baseURL = `http://${host}:${port}/${fixtureName}`;
        config.theme = {
          ...config.theme,
          preset: theme,
          navStyle,
          navActiveStyle: activeStyle
        };
        config.build.outputDir = "_site";
        writeFileSync(configPath, `${JSON.stringify(config, null, 2)}\n`);

        run(binaryPath, ["build"], { cwd: sourceRoot });
        cpSync(join(sourceRoot, "_site"), join(publicRoot, fixtureName), {
          recursive: true
        });
      }
    }
  }

  // The fixture deliberately includes a domain-root-relative Markdown image.
  // Each generated garden otherwise lives under a synthetic path so all
  // variants can share one test server.
  cpSync(join(publicRoot, "classic-base-base", "static"), join(publicRoot, "static"), {
    recursive: true
  });
}

const contentTypes = new Map([
  [".css", "text/css; charset=utf-8"],
  [".html", "text/html; charset=utf-8"],
  [".ico", "image/x-icon"],
  [".js", "text/javascript; charset=utf-8"],
  [".json", "application/json; charset=utf-8"],
  [".png", "image/png"],
  [".svg", "image/svg+xml; charset=utf-8"],
  [".woff", "font/woff"],
  [".woff2", "font/woff2"],
  [".xml", "application/xml; charset=utf-8"]
]);

function resolveRequestPath(pathname) {
  const decoded = decodeURIComponent(pathname);
  let target = resolve(publicRoot, `.${decoded}`);
  if (target !== publicRoot && !target.startsWith(`${publicRoot}${sep}`)) {
    return null;
  }
  if (existsSync(target) && statSync(target).isDirectory()) {
    target = join(target, "index.html");
  }
  return target;
}

buildFixtures();

const server = createServer((request, response) => {
  const url = new URL(request.url ?? "/", `http://${host}:${port}`);
  if (url.pathname === "/health") {
    response.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
    response.end("ok");
    return;
  }

  let filePath;
  try {
    filePath = resolveRequestPath(url.pathname);
  } catch {
    filePath = null;
  }
  if (!filePath || !existsSync(filePath) || !statSync(filePath).isFile()) {
    response.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    response.end("not found");
    return;
  }

  response.writeHead(200, {
    "cache-control": "no-store",
    "content-type": contentTypes.get(extname(filePath)) ?? "application/octet-stream"
  });
  response.end(readFileSync(filePath));
});

function shutDown() {
  server.close(() => {
    rmSync(workRoot, { recursive: true, force: true });
    process.exit(0);
  });
}

process.on("SIGINT", shutDown);
process.on("SIGTERM", shutDown);

server.listen(port, host, () => {
  console.log(`Theme conformance garden ready at http://${host}:${port}`);
});
