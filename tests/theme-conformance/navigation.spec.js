import { expect, test } from "@playwright/test";

const themes = ["classic", "aurora", "paper", "terminal"];
test.use({ hasTouch: true });

async function expectHeadingClear(page) {
  await expect.poll(() => page.locator("#lists-and-tasks").evaluate((heading) => {
    const nav = document.querySelector(".lp-nav");
    const pinned = ["fixed", "sticky"].includes(getComputedStyle(nav).position);
    return heading.getBoundingClientRect().top - (pinned ? nav.getBoundingClientRect().bottom : 0);
  })).toBeGreaterThanOrEqual(12);
}

for (const theme of themes) {
  test(`${theme} keeps RSS accessible on mobile without JavaScript`, async ({ browser, baseURL }) => {
    const context = await browser.newContext({
      baseURL,
      javaScriptEnabled: false,
      hasTouch: true,
      viewport: { width: 320, height: 844 }
    });
    const page = await context.newPage();

    try {
      await page.goto(`/${theme}-glassy-box/notes/components/`);
      await expect(page.locator("html")).toHaveClass(/\blp-no-js\b/);
      await expect(page.getByRole("button", { name: "Site menu", exact: true })).toBeHidden();
      await expect(page.getByRole("link", { name: "RSS feed", exact: true })).toBeVisible();
      await expect(page.getByRole("link", { name: "RSS feed", exact: true })).toHaveAttribute(
        "href",
        `/${theme}-glassy-box/feed.xml`
      );
    } finally {
      await context.close();
    }
  });

  for (const navStyle of ["base", "sticky", "glassy"]) {
    test(`${theme} ${navStyle} mobile navigation stays compact and headings remain clear`, async ({ page }) => {
      await page.setViewportSize({ width: 320, height: 844 });
      const url = `/${theme}-${navStyle}-box/notes/components/`;
      await page.goto(`${url}#lists-and-tasks`);
      await page.evaluate(() => document.fonts.ready);
      await expectHeadingClear(page);

      await page.locator(".lp-nav-title").evaluate((el) => {
        el.textContent = "The Personal Knowledge Garden of Alex Morgan";
      });
      await page.locator(".lp-nav-links").evaluate((el) => {
        el.children[1].textContent = "Research and Reading Notes";
        for (const label of ["Experiments", "Reading List", "Travel Journal"]) {
          const link = el.firstElementChild.cloneNode(true);
          link.textContent = label;
          el.append(link);
        }
      });

      for (const width of [390, 320]) {
        await page.setViewportSize({ width, height: 844 });
        await page.evaluate(() => scrollTo(0, 500));
        const nav = page.locator(".lp-nav");
        if (navStyle === "glassy") await expect(nav).toHaveClass(/lp-nav--pill/);
        if (navStyle === "glassy") {
          const box = await nav.boundingBox();
          expect(box.width).toBeLessThanOrEqual(Math.min(384, width - 32));
          expect(box.x + box.width / 2).toBeCloseTo(width / 2, 0);
        }
        if (theme === "classic" && navStyle !== "base") {
          await expect(nav).not.toHaveCSS("background-color", "rgba(0, 0, 0, 0)");
        }
        await expect.poll(() => nav.evaluate((el) => el.offsetHeight)).toBeLessThanOrEqual(navStyle === "glassy" ? 64 : 132);
        const layout = await nav.evaluate((el) => {
          const title = el.querySelector(".lp-nav-title");
          return {
            titleHeight: title.offsetHeight,
            titleLineHeight: parseFloat(getComputedStyle(title).lineHeight),
            titleOverflow: title.scrollWidth > title.clientWidth,
            navOverflow: el.scrollWidth - el.clientWidth,
            buttons: [...el.querySelectorAll(".lp-search-toggle, .lp-nav-menu-toggle")].map((button) => ({
              width: button.offsetWidth, height: button.offsetHeight
            }))
          };
        });
        expect(layout.titleHeight).toBeLessThanOrEqual(Math.ceil(layout.titleLineHeight));
        expect(layout.titleOverflow).toBe(true);
        expect(layout.navOverflow).toBeLessThanOrEqual(1);
        for (const button of layout.buttons) {
          expect(button.width).toBeGreaterThanOrEqual(44);
          expect(button.height).toBeGreaterThanOrEqual(44);
        }
        if (navStyle === "glassy") {
          await expect(page.locator(".lp-nav-links")).toBeHidden();
          await page.getByRole("button", { name: "Site menu", exact: true }).click();
          const menuLinks = page.locator(".lp-nav-menu-link");
          await expect(menuLinks).toHaveCount(7);
          await expect(menuLinks.filter({ hasText: "Research and Reading Notes" })).toHaveAttribute("aria-current", "page");
          await menuLinks.last().focus();
          await expect(menuLinks.last()).toBeInViewport();
          await page.keyboard.press("Escape");
        } else {
          await page.locator(".lp-nav-link").last().focus();
          await expect.poll(() => page.locator(".lp-nav-links").evaluate((el) => el.scrollLeft)).toBeGreaterThan(0);
        }
        await page.locator('#lists-and-tasks .lp-heading-anchor').evaluate((el) => el.click());
        await expectHeadingClear(page);
      }

      if (navStyle === "glassy") {
        // A size change without a viewport event must also refresh the offset.
        const oldOffset = await page.evaluate(() => parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--lp-nav-offset')));
        await page.locator('.lp-nav').evaluate((el) => { el.style.paddingTop = '24px'; });
        await expect.poll(() => page.evaluate(() => parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--lp-nav-offset')))).toBeGreaterThan(oldOffset);
        await page.locator('#lists-and-tasks .lp-heading-anchor').evaluate((el) => el.click());
        await expectHeadingClear(page);
        // Crossing the desktop breakpoint changes height and the normal-flow width.
        for (const width of [1024, 320]) {
          await page.setViewportSize({ width, height: 844 });
          await page.evaluate(() => scrollTo(0, 500));
          await expect(page.locator(".lp-nav")).toHaveClass(/lp-nav--pill/);
          await page.waitForTimeout(400);
          const floatedTop = await page.locator(".lp-main").evaluate((el) => el.getBoundingClientRect().top + scrollY);
          await page.evaluate(() => scrollTo(0, 0));
          await expect(page.locator(".lp-nav")).not.toHaveClass(/lp-nav--pill/);
          await expect.poll(() => page.locator(".lp-main").evaluate((el, top) =>
            Math.abs(el.getBoundingClientRect().top + scrollY - top), floatedTop)).toBeLessThanOrEqual(1);
        }
      }
    });
  }

  test(`${theme} mobile reader tools support touch, keyboard, and overlays`, async ({ page, browserName }) => {
    await page.setViewportSize({ width: 320, height: 844 });
    await page.goto(`/${theme}-glassy-box/notes/components/`);
    await page.evaluate(() => scrollTo(0, 500));
    const toggle = page.getByRole("button", { name: "Site menu", exact: true });
    const tools = page.locator(".lp-nav-tools");
    await expect(tools).toBeHidden();
    await toggle.tap();
    await expect(toggle).toHaveAttribute("aria-expanded", "true");
    await expect(tools).toBeVisible();
    for (const el of await tools.locator("a, button").all()) {
      expect((await el.boundingBox()).height).toBeGreaterThanOrEqual(44);
    }
    await toggle.focus();
    // macOS WebKit uses Option-Tab to include links in keyboard navigation.
    await page.keyboard.press(browserName === "webkit" && process.platform === "darwin" ? "Alt+Tab" : "Tab");
    await expect(page.locator(".lp-nav-menu-link").first()).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(tools).toBeHidden();
    await expect(toggle).toBeFocused();

    await toggle.click();
    await page.locator(".lp-theme-toggle").click();
    await expect(page.locator("html")).toHaveAttribute("data-theme-preference", "dark");
    await expect(tools).toBeHidden();
    await expect(toggle).toBeFocused();
    await toggle.click();
    await page.locator(".lp-graph-toggle").click();
    await expect(page.locator(".lp-graph-overlay")).toHaveClass(/lp-graph-overlay--open/);
    await expect(tools).toBeHidden();
    await page.keyboard.press("Escape");
    await expect(toggle).toBeFocused();

    await page.locator(".lp-search-toggle").click();
    await expect(page.locator(".lp-search-input")).toBeFocused();
    await page.keyboard.press("Escape");
    await toggle.click();
    await page.mouse.click(5, 800);
    await expect(tools).toBeHidden();
    await page.setViewportSize({ width: 1280, height: 800 });
    await expect(toggle).toBeHidden();
    await expect(page.locator(".lp-theme-toggle")).toBeVisible();
    await expect(page.locator(".lp-rss-link")).toBeVisible();
  });
}
