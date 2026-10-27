import { expect, test } from '@playwright/test';
const themes = ['classic', 'aurora', 'paper', 'quiet', 'terminal'];
for (const theme of themes) {
  test(`${theme} Everything adapts to width and preserves tools and destinations`, async ({ page }) => {
    const errors = [];
    page.on('pageerror', e => errors.push(e.message));
    await page.setViewportSize({width: 1440, height: 900});
    await page.goto(`/${theme}-everything/notes/components/`);
    await page.evaluate(() => document.fonts.ready);
    await expect.poll(() => page.locator('.lp-nav-container').evaluate(el => getComputedStyle(el).gridTemplateRows.split(' ').length)).toBe(2);
    const picker = page.locator('.lp-section-picker');
    const summary = picker.locator('summary');
    const shortcuts = page.locator('.lp-section-nav > .lp-nav-links');
    await expect(shortcuts.locator('a:visible').first()).toHaveText('Notes');
    for (const width of [320, 390, 600, 1440, 320]) {
      await page.setViewportSize({width, height: 900});
      await expect.poll(() => page.locator('.lp-nav').evaluate(nav => [...nav.querySelectorAll('a, button, summary')].filter(e => e.getClientRects().length).every(e => { const r = e.getBoundingClientRect(); return r.left >= 0 && r.right <= innerWidth + 1; }))).toBe(true);
      await expect.poll(() => shortcuts.evaluate(el => [...el.children].filter(a => !a.hidden).every(a => a.getBoundingClientRect().right <= el.closest('.lp-section-nav').getBoundingClientRect().right + 1))).toBe(true);
      await summary.click();
      await expect(picker).toHaveAttribute('open', '');
      await expect(picker.getByRole('link')).toHaveCount(5);
      await expect(picker.getByRole('link', {name:'External & more'})).toHaveAttribute('href', 'https://example.com/?a=1&b=2');
      await page.keyboard.press('Escape');
      await expect(picker).not.toHaveAttribute('open');
      await expect(summary).toBeFocused();
      if (width <= 420) await page.getByRole('button', {name:'Site menu', exact:true}).click();
      await expect(page.getByRole('link', {name:'RSS feed', exact:true})).toBeVisible();
      await expect(page.getByRole('button', {name:'Open knowledge graph',exact:true})).toBeVisible();
      await page.keyboard.press('Escape');
      await page.getByRole('button', {name:'Search',exact:true}).click();
      await expect(page.locator('.lp-search-input')).toBeFocused();
      await page.keyboard.press('Escape');
    }
    await expect(summary).toBeVisible();
    expect(errors).toEqual([]);
  });
  test(`${theme} Everything remains usable without JavaScript`, async ({ browser, baseURL }) => {
    const context = await browser.newContext({baseURL, javaScriptEnabled:false, viewport:{width:320,height:844}});
    try {
      const page = await context.newPage();
      await page.goto(`/${theme}-everything/notes/components/`);
      await page.locator('.lp-section-picker summary').click();
      await expect(page.locator('.lp-section-menu a')).toHaveCount(5);
      await page.locator('.lp-section-menu').getByRole('link',{name:'Everything',exact:true}).click();
      await expect(page.locator('.lp-section-title')).toHaveText('Everything');
      await expect(page.getByRole('link',{name:'RSS feed',exact:true})).toBeVisible();
    } finally { await context.close(); }
  });
}
