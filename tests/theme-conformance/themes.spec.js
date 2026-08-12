import { expect, test } from "@playwright/test";

const themes = ["classic", "aurora"];
const navStyles = ["base", "sticky", "glassy"];
const activeStyles = ["base", "underlined", "box"];
const colorSchemes = ["light", "dark"];
const viewports = {
  desktop: { width: 1280, height: 800 },
  mobile: { width: 390, height: 844 }
};

function fixtureName(theme, navStyle, activeStyle) {
  return `${theme}-${navStyle}-${activeStyle}`;
}

async function numericFontSize(locator) {
  return locator.evaluate((element) => Number.parseFloat(getComputedStyle(element).fontSize));
}

for (const theme of themes) {
  for (const navStyle of navStyles) {
    for (const activeStyle of activeStyles) {
      for (const [viewportName, viewport] of Object.entries(viewports)) {
        for (const colorScheme of colorSchemes) {
          const state = `${theme} / ${navStyle} / ${activeStyle} / ${viewportName} / ${colorScheme}`;

          test(state, async ({ page }) => {
            const failedResponses = [];
            page.on("response", (response) => {
              if (response.status() >= 400 && response.url().startsWith("http://127.0.0.1")) {
                failedResponses.push(`${response.status()} ${response.url()}`);
              }
            });

            await page.setViewportSize(viewport);
            await page.emulateMedia({ colorScheme });
            await page.goto(`/${fixtureName(theme, navStyle, activeStyle)}/notes/components/`);
            await page.evaluate(() => document.fonts.ready);

            const root = page.locator("html");
            await expect(root).toHaveAttribute("data-lp-theme", theme);
            await expect(root).toHaveAttribute("data-theme", colorScheme);
            await expect(page.locator(".lp-title")).toHaveText("Component Gallery");

            const activeLink = page.locator(".lp-nav-link--active");
            await expect(activeLink).toHaveCount(1);
            await expect(activeLink).toHaveClass(new RegExp(`\\blp-nav-active-${activeStyle}\\b`));

            const activeTreatment = await activeLink.evaluate((element) => {
              const computed = getComputedStyle(element);
              const after = getComputedStyle(element, "::after");
              const accent = getComputedStyle(document.documentElement)
                .getPropertyValue("--lp-accent")
                .trim();
              const probe = document.createElement("span");
              probe.style.color = accent;
              document.body.append(probe);
              const accentColor = getComputedStyle(probe).color;
              probe.remove();
              return {
                backgroundColor: computed.backgroundColor,
                backgroundImage: computed.backgroundImage,
                boxShadow: computed.boxShadow,
                color: computed.color,
                afterContent: after.content,
                afterHeight: Number.parseFloat(after.height) || 0,
                accent,
                accentColor
              };
            });

            if (activeStyle === "base") {
              expect(activeTreatment.accent).not.toBe("");
              expect(activeTreatment.color).toBe(activeTreatment.accentColor);
            } else if (activeStyle === "underlined") {
              expect(
                activeTreatment.boxShadow !== "none" ||
                  (activeTreatment.afterContent !== "none" && activeTreatment.afterHeight >= 1)
              ).toBe(true);
            } else {
              expect(
                activeTreatment.backgroundColor !== "rgba(0, 0, 0, 0)" ||
                  activeTreatment.backgroundImage !== "none"
              ).toBe(true);
            }

            const nav = page.locator(".lp-nav");
            if (navStyle === "base") {
              await expect(nav).toHaveCSS("position", "static");
            } else if (navStyle === "sticky") {
              await expect(nav).toHaveCSS("position", "sticky");
            } else {
              await page.evaluate(() => window.scrollTo(0, 500));
              await expect(nav).toHaveClass(/\blp-nav--pill\b/);
              await expect(nav).toHaveCSS("position", "fixed");
              await expect
                .poll(() => nav.evaluate((element) => element.getBoundingClientRect().top))
                .toBeGreaterThanOrEqual(0);
              expect(
                await nav.evaluate((element) => element.getBoundingClientRect().top)
              ).toBeLessThanOrEqual(16);
            }

            expect(await numericFontSize(page.locator(".lp-body"))).toBeCloseTo(16, 1);
            expect(await numericFontSize(page.locator(".lp-title"))).toBeCloseTo(
              viewportName === "mobile" ? 24 : 32,
              1
            );
            expect(await numericFontSize(page.locator(".lp-content h1").first())).toBeCloseTo(28, 1);
            expect(await numericFontSize(page.locator(".lp-content h2").first())).toBeCloseTo(24, 1);
            expect(await numericFontSize(page.locator(".lp-content h3").first())).toBeCloseTo(20, 1);

            const horizontalOverflow = await page.evaluate(
              () => document.documentElement.scrollWidth - document.documentElement.clientWidth
            );
            expect(horizontalOverflow).toBeLessThanOrEqual(1);

            if (viewportName === "mobile") {
              const alignment = await page.evaluate(() => {
                const title = document.querySelector(".lp-nav-title");
                const firstLink = document.querySelector(".lp-nav-link");
                const textStart = (element) => {
                  const textNode = [...element.childNodes].find(
                    (node) => node.nodeType === Node.TEXT_NODE && node.textContent.trim() !== ""
                  );
                  const range = document.createRange();
                  range.setStart(textNode, 0);
                  range.setEnd(textNode, 1);
                  return range.getBoundingClientRect().left;
                };
                return Math.abs(textStart(title) - textStart(firstLink));
              });
              expect(alignment).toBeLessThanOrEqual(1);
            }

            expect(failedResponses).toEqual([]);
          });
        }
      }
    }
  }
}

for (const theme of themes) {
  test(`${theme} exposes the full fixture and reader tools`, async ({ page }) => {
    await page.setViewportSize(viewports.desktop);
    await page.emulateMedia({ colorScheme: "light" });
    const fixture = fixtureName(theme, "base", "base");
    await page.goto(`/${fixture}/notes/components/`);

    await expect(page.locator(".lp-article")).toBeVisible();
    await expect(page.locator(".lp-toc")).toBeVisible();
    await expect(page.locator(".lp-content table")).toBeVisible();
    await expect(page.locator(".lp-content pre").first()).toBeVisible();
    await expect(page.locator(".lp-content blockquote")).toBeVisible();
    await expect(page.locator(".lp-backlinks")).toBeVisible();

    await page.locator(".lp-theme-toggle").click();
    await expect(page.locator("html")).toHaveAttribute("data-theme-preference", "dark");
    await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

    const searchIndexLoaded = page.waitForResponse(
      (response) => response.url().endsWith("/search-index.json") && response.ok()
    );
    await page.locator(".lp-search-toggle").click();
    await expect(page.locator(".lp-search-overlay")).toHaveClass(/\blp-search-overlay--open\b/);
    await expect(page.locator(".lp-search-input")).toBeFocused();
    await searchIndexLoaded;
    await page.locator(".lp-search-input").fill("callout");
    await expect(page.locator(".lp-search-result").first()).toBeVisible();
    await page.keyboard.press("Escape");

    await page.locator(".lp-graph-toggle").click();
    await expect(page.locator(".lp-graph-overlay")).toHaveClass(/\blp-graph-overlay--open\b/);
    await expect(page.locator(".lp-graph-node").first()).toBeVisible();
    await page.keyboard.press("Escape");

    await page.goto(`/${fixture}/notes/callouts/`);
    expect(await page.locator(".lp-callout").count()).toBeGreaterThan(5);

    await page.goto(`/${fixture}/notes/`);
    await expect(page.locator(".lp-section")).toBeVisible();
    await expect(page.locator(".lp-index-item").first()).toBeVisible();

    await page.goto(`/${fixture}/tags/`);
    await expect(page.locator(".lp-tag-cloud-item").first()).toBeVisible();
  });
}
