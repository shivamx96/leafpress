import { expect, test } from "@playwright/test";

// Keep this suite small enough to run on secondary browser engines on every PR.
for (const theme of ["classic", "aurora", "paper", "terminal"]) {
  test(`${theme} supports essential reader flows`, async ({ page }) => {
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.emulateMedia({ colorScheme: "light" });
    await page.goto(`/${theme}-base-base/notes/components/`);
    await expect(page.locator(".lp-article")).toBeVisible();
    await expect(page.locator(".lp-toc")).toBeVisible();

    await page.locator(".lp-theme-toggle").click();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

    const indexLoaded = page.waitForResponse(
      (response) => response.url().endsWith("/search-index.json") && response.ok()
    );
    await page.locator(".lp-search-toggle").click();
    await indexLoaded;
    await expect(page.locator(".lp-search-input")).toBeFocused();
    await page.locator(".lp-search-input").fill("callout");
    await expect(page.locator(".lp-search-result").first()).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.locator(".lp-search-overlay")).not.toHaveClass(/lp-search-overlay--open/);

    await page.locator(".lp-graph-toggle").click();
    await expect(page.locator(".lp-graph-node").first()).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.locator(".lp-graph-overlay")).not.toHaveClass(/lp-graph-overlay--open/);
    expect(errors).toEqual([]);
  });
}
