import { expect, test } from '@playwright/test';
const themes = ['classic', 'aurora', 'paper', 'quiet', 'terminal'];
for (const theme of themes) {
  for (const colorScheme of ['light', 'dark']) {
    test(`${theme} bottom dock in ${colorScheme} preserves reading and opens upward`, async ({ page }) => {
      await page.emulateMedia({ colorScheme });
      const errors = [];
      page.on('pageerror', e => errors.push(e.message));
      for (const width of [1440, 600, 390, 320]) {
        await page.setViewportSize({ width, height: 844 });
        await page.goto(`/${theme}-everything-bottom/notes/components/#lists-and-tasks`);
        await page.evaluate(() => document.fonts.ready);
        const nav = page.locator('.lp-nav');
        await expect(nav).toHaveCSS('position', 'fixed');
        await expect.poll(() => nav.evaluate(el => {
          const r = el.getBoundingClientRect();
          return r.left >= 0 && r.right <= innerWidth && r.bottom <= innerHeight && r.top > innerHeight / 2 && r.height <= 80;
        })).toBe(true);
        await expect.poll(() => page.locator('#lists-and-tasks').evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(0);
        const summary = page.locator('.lp-section-picker summary');
        await summary.click();
        const menu = page.locator('.lp-section-menu');
        await expect(menu).toBeVisible();
        await expect.poll(() => menu.evaluate(el => {
          const r=el.getBoundingClientRect(), nav=document.querySelector('.lp-nav').getBoundingClientRect();
          return r.bottom <= nav.top && r.top >= 0 && r.left >= 0 && r.right <= innerWidth;
        })).toBe(true);
        await expect(menu.locator('a[href$="/notes/"]')).toHaveAttribute('aria-current','page');
        await page.keyboard.press('Escape');
        await expect(summary).toBeFocused();
        if (width <= 420) {
          await page.getByRole('button',{name:'Site menu',exact:true}).click();
          await expect(page.locator('.lp-nav-tools')).toBeVisible();
          await expect.poll(() => page.locator('.lp-nav-tools').evaluate(el => el.getBoundingClientRect().bottom <= document.querySelector('.lp-nav').getBoundingClientRect().top)).toBe(true);
        }
        await page.getByRole('button',{name:'Open knowledge graph',exact:true}).click();
        await expect(page.locator('.lp-graph-overlay')).toBeVisible();
        await page.keyboard.press('Escape');
        await page.getByRole('button',{name:'Search',exact:true}).click();
        await expect(page.locator('.lp-search-input')).toBeFocused();
        await page.keyboard.press('Escape');
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        await expect.poll(() => page.locator('.lp-footer').evaluate(el => el.getBoundingClientRect().bottom <= document.querySelector('.lp-nav').getBoundingClientRect().top)).toBe(true);
      }
      expect(errors).toEqual([]);
    });
  }
  test(`${theme} bottom supports legacy links and no-JS section picker`, async ({browser, baseURL}) => {
    const context = await browser.newContext({ baseURL, javaScriptEnabled:false, viewport:{width:320,height:844} });
    try {
      const page = await context.newPage();
      for (const fixture of [`${theme}-bottom-box`,`${theme}-everything-bottom`]) {
        await page.goto(`/${fixture}/notes/components/`);
        await expect(page.locator('.lp-site-title')).toBeVisible();
        await expect(page.getByRole('link',{name:'RSS feed',exact:true})).toBeVisible();
        if (fixture.includes('everything')) {
          await page.locator('.lp-section-picker summary').click();
          await page.locator('.lp-section-menu').getByRole('link',{name:'Everything',exact:true}).click();
          await expect(page.locator('.lp-section-title')).toHaveText('Everything');
        } else {
          await expect(page.locator('.lp-nav-links a').first()).toBeVisible();
        }
      }
    } finally { await context.close(); }
  });
}
