import { expect, test } from "@playwright/test";

for (const system of ["light", "dark"]) {
  test(`theme toggle follows and overrides a ${system} system preference`, async ({ page }) => {
    const opposite = system === "light" ? "dark" : "light";
    const root = page.locator("html");
    const toggle = page.locator(".lp-theme-toggle");
    const storedTheme = () => page.evaluate(() => localStorage.getItem("theme"));
    await page.emulateMedia({ colorScheme: system });
    await page.goto("/classic-base-base/notes/components/");
    await expect(root).toHaveAttribute("data-theme-preference", "system");
    await expect(root).toHaveAttribute("data-theme", system);
    await expect(toggle).toHaveAccessibleName(`Switch to ${opposite} theme`);

    await toggle.click();
    await expect(root).toHaveAttribute("data-theme", opposite);
    expect(await storedTheme()).toBe(opposite);
    await page.reload();
    await expect(root).toHaveAttribute("data-theme", opposite);

    // OS changes, including matching the override, must not discard it.
    await page.emulateMedia({ colorScheme: opposite });
    await expect(root).toHaveAttribute("data-theme-preference", opposite);
    await page.emulateMedia({ colorScheme: system });
    await expect(root).toHaveAttribute("data-theme", opposite);
    expect(await storedTheme()).toBe(opposite);

    await toggle.click();
    await expect(root).toHaveAttribute("data-theme", system);
    await expect(root).toHaveAttribute("data-theme-preference", "system");
    expect(await storedTheme()).toBeNull();
    await page.reload();
    await page.emulateMedia({ colorScheme: opposite });
    await expect(root).toHaveAttribute("data-theme", opposite);

    // A saved override matching the OS still switches appearance on click.
    await page.evaluate((theme) => localStorage.setItem("theme", theme), opposite);
    await page.reload();
    await toggle.click();
    await expect(root).toHaveAttribute("data-theme", system);
    await expect(root).toHaveAttribute("data-theme-preference", system);
    expect(await storedTheme()).toBe(system);
  });
}

// Keep this suite small enough to run on secondary browser engines on every PR.
for (const theme of ["classic", "aurora", "paper", "quiet", "terminal"]) {
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

for (const layout of ["classic-everything", "classic-everything-bottom"]) {
  for (const system of ["light", "dark"]) {
    test(`${layout} compact theme label reflects preference with ${system} system`, async ({ page }) => {
      const opposite = system === "light" ? "dark" : "light";
      const stateName = opposite === "dark" ? "Dark" : "Light";
      await page.setViewportSize({ width: 375, height: 812 });
      await page.emulateMedia({ colorScheme: system });
      await page.goto(`/${layout}/notes/components/`);
      const toggle = page.locator(".lp-theme-toggle");
      const label = toggle.locator(".lp-nav-tool-label");
      const openMenu = () => page.getByRole("button", { name: "Site menu", exact: true }).click();
      await openMenu();
      await expect(label).toBeVisible();
      await expect(label).toHaveText("Auto");
      await toggle.click();
      await expect(label).toHaveText(stateName);
      await page.reload();
      await openMenu();
      await expect(label).toHaveText(stateName);
      // An OS change must not turn an explicit override into Auto.
      await page.emulateMedia({ colorScheme: opposite });
      await expect(label).toHaveText(stateName);
      await page.emulateMedia({ colorScheme: system });
      await toggle.click();
      await expect(label).toHaveText("Auto");
      await page.emulateMedia({ colorScheme: opposite });
      await expect(label).toHaveText("Auto");
      await expect(page.locator("html")).toHaveAttribute("data-theme", opposite);
    });
  }
}
