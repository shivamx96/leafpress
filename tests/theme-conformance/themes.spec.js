import { expect, test } from "@playwright/test";

const themes = ["classic", "aurora", "paper", "terminal"];
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

async function followLastFootnoteBack(page) {
  const backref = page.locator('.footnote-backref[role="doc-backlink"]').last();
  await backref.scrollIntoViewIfNeeded();
  const referenceHref = await backref.getAttribute("href");
  expect(referenceHref).toMatch(/^#fnref:/);

  const reference = page.locator(`[id="${referenceHref.slice(1)}"]`);
  await backref.click();
  await expect(page).toHaveURL(new RegExp(`${referenceHref.replace(":", "\\:")}$`));
  await expect(reference).toBeInViewport();

  const placement = await reference.evaluate((element) => {
    const target = element.getBoundingClientRect();
    const nav = document.querySelector(".lp-nav");
    const navPosition = getComputedStyle(nav).position;
    const navBottom = ["fixed", "sticky"].includes(navPosition)
      ? nav.getBoundingClientRect().bottom
      : 0;
    return {
      clearance: target.top - navBottom,
      targetBottom: target.bottom,
      viewportHeight: window.innerHeight
    };
  });
  expect(placement.clearance).toBeGreaterThanOrEqual(8);
  expect(placement.targetBottom).toBeLessThanOrEqual(placement.viewportHeight);
}

for (const theme of themes) {
  test(`${theme} titles use the available content width`, async ({ page }) => {
    await page.setViewportSize(viewports.desktop);
    await page.goto(`/${fixtureName(theme, "base", "base")}/notes/components/`);
    await page.evaluate(() => document.fonts.ready);

    const title = page.locator(".lp-title");
    await title.evaluate((element) => {
      element.textContent = "A Practical Guide to Connected Systems";
    });

    const layout = await title.evaluate((element) => {
      const style = getComputedStyle(element);
      return {
        availableWidth: element.parentElement.getBoundingClientRect().width,
        lineHeight: Number.parseFloat(style.lineHeight),
        height: element.getBoundingClientRect().height,
        width: element.getBoundingClientRect().width
      };
    });

    expect(layout.width).toBeCloseTo(layout.availableWidth, 0);
    expect(layout.height).toBeLessThan(layout.lineHeight * 1.5);
  });

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

              if (["paper", "terminal"].includes(theme)) {
                const inactiveBox = await page
                  .locator(".lp-nav-link:not(.lp-nav-link--active)")
                  .first()
                  .boundingBox();
                const activeBox = await activeLink.boundingBox();
                expect(activeBox).not.toBeNull();
                expect(inactiveBox).not.toBeNull();
                expect(activeBox.height).toBeCloseTo(inactiveBox.height, 1);
                expect(activeBox.y + activeBox.height / 2).toBeCloseTo(
                  inactiveBox.y + inactiveBox.height / 2,
                  1
                );
              }
            }

            const nav = page.locator(".lp-nav");
            if (navStyle === "base") {
              await expect(nav).toHaveCSS("position", "static");
            } else if (navStyle === "sticky") {
              await expect(nav).toHaveCSS("position", "sticky");
            } else {
              const initialBox = theme === "paper" ? await nav.boundingBox() : null;
              if (theme === "paper") {
                await expect(nav).toHaveCSS("position", "fixed");
                expect(initialBox).not.toBeNull();
                expect(initialBox.x).toBeGreaterThan(0);
                expect(initialBox.y).toBeGreaterThan(0);
                expect(initialBox.width).toBeLessThan(viewport.width);
              }

              await page.evaluate(() => window.scrollTo(0, 500));
              await expect(nav).toHaveClass(/\blp-nav--pill\b/);
              await expect(nav).toHaveCSS("position", "fixed");
              await expect
                .poll(() => nav.evaluate((element) => element.getBoundingClientRect().top))
                .toBeGreaterThanOrEqual(0);
              expect(
                await nav.evaluate((element) => element.getBoundingClientRect().top)
              ).toBeLessThanOrEqual(16);

              if (theme === "paper") {
                const scrolledBox = await nav.boundingBox();
                expect(scrolledBox.x).toBeCloseTo(initialBox.x, 0);
                expect(scrolledBox.y).toBeCloseTo(initialBox.y, 0);
                expect(scrolledBox.width).toBeCloseTo(initialBox.width, 0);
              }
            }

            expect(await numericFontSize(page.locator(".lp-body"))).toBeCloseTo(16, 1);
            expect(await numericFontSize(page.locator(".lp-title"))).toBeCloseTo(
              viewportName === "mobile" ? 24 : 32,
              1
            );
            expect(await numericFontSize(page.locator(".lp-content h1").first())).toBeCloseTo(28, 1);
            expect(await numericFontSize(page.locator(".lp-content h2").first())).toBeCloseTo(24, 1);
            expect(await numericFontSize(page.locator(".lp-content h3").first())).toBeCloseTo(20, 1);

            const footnoteReference = page.locator(".footnote-ref").first();
            const footnotes = page.locator('.footnotes[role="doc-endnotes"]');
            await expect(footnoteReference).toBeVisible();
            await expect(footnoteReference).toHaveAttribute("role", "doc-noteref");
            await expect(footnotes).toBeVisible();
            expect(await numericFontSize(footnotes)).toBeCloseTo(14, 1);
            await expect(footnotes.locator("li")).toHaveCount(2);
            await expect(footnotes.locator('.footnote-backref[role="doc-backlink"]')).toHaveCount(2);
            await followLastFootnoteBack(page);

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
    await expect(page.locator(".footnote-ref")).toHaveCount(2);
    await expect(page.locator(".footnotes")).toBeVisible();
    await expect(page.locator(".footnote-backref")).toHaveCount(2);
    await expect(page.locator(".lp-backlinks")).toBeVisible();

    const lastFootnote = page.locator(".footnote-ref").last();
    const footnoteTarget = await lastFootnote.getAttribute("href");
    expect(footnoteTarget).toMatch(/^#fn:/);
    await lastFootnote.click();
    await expect(page).toHaveURL(new RegExp(`${footnoteTarget.replace(":", "\\:")}$`));
    await followLastFootnoteBack(page);

    if (theme === "terminal") {
      const terminalChrome = await page.evaluate(() => ({
        article: getComputedStyle(document.querySelector(".lp-article"), "::before").content,
        title: getComputedStyle(document.querySelector(".lp-title"), "::before").content,
        titleAfter: getComputedStyle(document.querySelector(".lp-title"), "::after").content,
        toc: getComputedStyle(document.querySelector(".lp-toc-nav"), "::before").content,
        metadata: getComputedStyle(document.querySelector(".lp-meta"), "::before").content,
        contentHeading: getComputedStyle(document.querySelector(".lp-content h2"), "::before").content,
        codeHeader: getComputedStyle(document.querySelector(".lp-content pre"), "::before").content,
        actionIcons: [...document.querySelectorAll(".lp-nav-actions svg")].map(
          (element) => getComputedStyle(element).display
        )
      }));
      expect(terminalChrome.article).toBe('"~/garden/notes"');
      expect(terminalChrome.title).toBe('"$ "');
      expect(terminalChrome.titleAfter).toBe("none");
      expect(terminalChrome.toc).toBe('"contents"');
      expect(terminalChrome.metadata).toBe("none");
      expect(terminalChrome.contentHeading).toBe("none");
      expect(terminalChrome.codeHeader).toBe("none");
      expect(terminalChrome.actionIcons.length).toBeGreaterThan(0);
      expect(terminalChrome.actionIcons.some((display) => display !== "none")).toBe(true);
    }

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
    if (theme === "terminal") {
      const searchCommand = await page
        .locator(".lp-search-header")
        .evaluate((element) => getComputedStyle(element, "::before").content);
      expect(searchCommand).toBe('"$"');
    }
    await page.keyboard.press("Escape");

    await page.locator(".lp-graph-toggle").click();
    await expect(page.locator(".lp-graph-overlay")).toHaveClass(/\blp-graph-overlay--open\b/);
    await expect(page.locator(".lp-graph-node").first()).toBeVisible();
    if (theme === "terminal") {
      const graphCommand = await page
        .locator(".lp-graph-panel")
        .evaluate((element) => getComputedStyle(element, "::before").content);
      expect(graphCommand).toBe('"graph"');
    }
    await page.keyboard.press("Escape");

    await page.locator(".lp-theme-toggle").click();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "light");

    await page.goto(`/${fixture}/notes/callouts/`);
    expect(await page.locator(".lp-callout").count()).toBeGreaterThan(5);

    if (theme === "terminal") {
      await expect(page.locator(".lp-callout-icon").first()).toHaveCSS("display", "none");
      const titlePresentation = await page.locator(".lp-callout-title").first().evaluate((element) => {
        const computed = getComputedStyle(element);
        return {
          clipPath: computed.clipPath,
          height: Number.parseFloat(computed.height),
          width: Number.parseFloat(computed.width)
        };
      });
      expect(titlePresentation.clipPath).not.toBe("none");
      expect(titlePresentation.height).toBe(1);
      expect(titlePresentation.width).toBe(1);

      const statusContrasts = await page.locator(".lp-callout").evaluateAll((callouts) => {
        const colorChannels = (value) => {
          const canvas = document.createElement("canvas");
          canvas.width = 1;
          canvas.height = 1;
          const context = canvas.getContext("2d");
          context.fillStyle = value;
          context.fillRect(0, 0, 1, 1);
          return [...context.getImageData(0, 0, 1, 1).data].slice(0, 3);
        };
        const luminance = (value) => {
          const channels = colorChannels(value).map((channel) => {
            const normalized = channel / 255;
            return normalized <= 0.04045
              ? normalized / 12.92
              : ((normalized + 0.055) / 1.055) ** 2.4;
          });
          return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
        };
        return callouts.map((callout) => {
          const foreground = luminance(getComputedStyle(callout, "::before").color);
          const background = luminance(getComputedStyle(callout).backgroundColor);
          return (Math.max(foreground, background) + 0.05) /
            (Math.min(foreground, background) + 0.05);
        });
      });
      expect(Math.min(...statusContrasts)).toBeGreaterThanOrEqual(4.5);
    }

    await page.goto(`/${fixture}/notes/`);
    await expect(page.locator(".lp-section")).toBeVisible();
    await expect(page.locator(".lp-index-item").first()).toBeVisible();
    const index = page.locator(".lp-index");
    await expect(index).toHaveClass(/\blp-index--columns-2\b/);
    for (const columns of [1, 2, 3]) {
      const renderedColumns = await index.evaluate((element, value) => {
        element.classList.remove(
          "lp-index--columns-1",
          "lp-index--columns-2",
          "lp-index--columns-3"
        );
        element.classList.add(`lp-index--columns-${value}`);
        return getComputedStyle(element).gridTemplateColumns.split(" ").length;
      }, columns);
      expect(renderedColumns).toBe(columns);
    }

    await page.setViewportSize(viewports.mobile);
    expect(
      await index.evaluate((element) =>
        getComputedStyle(element).gridTemplateColumns.split(" ").length
      )
    ).toBe(1);
    await page.setViewportSize(viewports.desktop);

    if (theme === "terminal") {
      const listChrome = await page.evaluate(() => ({
        permissionPrefix: getComputedStyle(document.querySelector(".lp-index-item"), "::before")
          .content,
        totalPrefix: getComputedStyle(document.querySelector(".lp-section-count"), "::before").content,
        footerStatus: getComputedStyle(document.querySelector(".lp-footer"), "::before").content
      }));
      expect(listChrome.permissionPrefix).toBe("none");
      expect(listChrome.totalPrefix).toBe("none");
      expect(listChrome.footerStatus).toBe("none");
    }

    await page.goto(`/${fixture}/tags/`);
    await expect(page.locator(".lp-tag-cloud-item").first()).toBeVisible();

    if (theme === "terminal") {
      await page.setViewportSize(viewports.mobile);
      await page.goto(`/${fixture}/notes/components/`);
      await page.locator(".lp-graph-toggle").click();
      const graphPanel = await page.locator(".lp-graph-panel").boundingBox();
      expect(graphPanel).not.toBeNull();
      expect(graphPanel.height).toBeLessThanOrEqual(viewports.mobile.height * 0.75);
      await page.keyboard.press("Escape");

      await page.goto(`/${fixture}/404.html`);
      await expect(page.locator(".lp-not-found-title")).toHaveText("404");
      await expect(page.locator(".lp-not-found-title")).toHaveCSS("font-size", "32px");
      const errorPrefix = await page
        .locator(".lp-not-found-title")
        .evaluate((element) => getComputedStyle(element, "::before").content);
      expect(errorPrefix).toBe('"ERR: "');
    }
  });
}
