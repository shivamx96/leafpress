import { expect, test } from '@playwright/test';
const themes = ['classic', 'aurora', 'paper', 'quiet', 'terminal'];
async function expectSymmetricIcons(page) {
  await expect.poll(() => page.locator('.lp-nav').evaluate(nav => {
    const home = nav.querySelector('.lp-dock-home');
    const group = nav.querySelector('.lp-dock-home-group').getBoundingClientRect();
    const actions = nav.querySelector('.lp-nav-actions').getBoundingClientRect();
    const controls = [...nav.querySelectorAll('.lp-nav-actions button, .lp-nav-actions .lp-rss-link')]
      .filter(el => el.getClientRects().length && getComputedStyle(el).visibility !== 'hidden')
      .sort((a, b) => a.getBoundingClientRect().left - b.getBoundingClientRect().left);
    const rect = home.getBoundingClientRect();
    if (!controls.length) return false;
    const first = controls[0].getBoundingClientRect();
    const last = controls[controls.length - 1].getBoundingClientRect();
    const dock = nav.getBoundingClientRect();
    const center = r => (r.left + r.right) / 2;
    return [home, ...controls].every(el => {
      const r = el.getBoundingClientRect();
      return Math.abs(r.width - 44) < 0.1 && Math.abs(r.height - 44) < 0.1;
    }) && Math.abs((center(rect) - dock.left) - (dock.right - center(last))) < 1
       && Math.abs((group.right - 0.5 - center(rect)) - (center(first) - actions.left - 0.5)) < 1;
  })).toBe(true);
}
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
        const home = nav.locator('.lp-dock-home');
        await expect(home).toHaveAttribute('href', `/${theme}-everything-bottom/`);
        await expect(home).toHaveAccessibleName('The Theme Garden — Home');
        await expect(page.locator('.lp-site-header')).toHaveCount(0);
        await expect(home.locator('svg')).toBeVisible();
        await expectSymmetricIcons(page);
        await expect(home.locator('.lp-dock-home-name')).toHaveCount(0);
        await expect.poll(() => home.evaluate(el => {
          const r = el.getBoundingClientRect(), picker = document.querySelector('.lp-section-picker summary').getBoundingClientRect();
          const actions = document.querySelector('.lp-nav-actions').getBoundingClientRect();
          return r.width >= 44 && r.height >= 44 && r.right <= picker.left && picker.right <= actions.left && actions.right <= innerWidth;
        })).toBe(true);
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
        await home.click();
        await expect(page).toHaveURL(new RegExp(`/${theme}-everything-bottom/$`));
        await expect(page.locator('.lp-dock-home')).toHaveAttribute('aria-current', 'page');
      }
      expect(errors).toEqual([]);
    });
  }
  test(`${theme} legacy dock keeps Home next to section links`, async ({page}) => {
    for (const width of [1440, 600, 320]) {
      await page.setViewportSize({width, height:844});
      await page.goto(`/${theme}-bottom-box/notes/components/`);
      await page.evaluate(() => document.fonts.ready);
      const home = page.locator('.lp-dock-home');
      await expect(home.locator('svg')).toBeVisible();
        await expectSymmetricIcons(page);
      await expect(page.locator('.lp-dock-home-group')).toHaveCSS('border-right-width', '1px');
      await expect.poll(() => home.evaluate(el => {
        const home = el.parentElement.getBoundingClientRect();
        const links = document.querySelector('.lp-nav-container > .lp-nav-links').getBoundingClientRect();
        return links.left - home.right;
      })).toBeLessThanOrEqual(12);
      await expect.poll(() => home.evaluate(el => {
        const home = el.parentElement.getBoundingClientRect();
        const links = document.querySelector('.lp-nav-container > .lp-nav-links').getBoundingClientRect();
        return links.left - home.right;
      })).toBeGreaterThanOrEqual(0);
      await expect(page.locator('.lp-nav-actions')).toHaveCSS('border-left-width', '1px');
    }
  });
  test(`${theme} bottom supports legacy links and no-JS section picker`, async ({browser, baseURL}) => {
    const context = await browser.newContext({ baseURL, javaScriptEnabled:false, viewport:{width:320,height:844} });
    try {
      const page = await context.newPage();
      for (const fixture of [`${theme}-bottom-box`,`${theme}-everything-bottom`]) {
        await page.goto(`/${fixture}/notes/components/`);
        await expect(page.locator('.lp-dock-home')).toBeVisible();
        await page.setViewportSize({width:1440, height:844});
        await expect(page.locator('.lp-dock-home-icon')).toBeVisible();
        await expect(page.locator('.lp-dock-home-name')).toHaveCount(0);
        await page.setViewportSize({width:320, height:844});
        await expect(page.locator('.lp-site-header')).toHaveCount(0);
        await expect(page.getByRole('link',{name:'RSS feed',exact:true})).toBeVisible();
        await expect(page.locator('.lp-nav-actions button:visible')).toHaveCount(0);
        if (fixture.includes('everything')) {
          await page.locator('.lp-section-picker summary').click();
          await page.locator('.lp-section-menu').getByRole('link',{name:'Everything',exact:true}).click();
          await expect(page.locator('.lp-section-title')).toHaveText('Everything');
        } else {
          await expect(page.locator('.lp-nav-links a').first()).toBeVisible();
        }
        await page.locator('.lp-dock-home').click();
        await expect(page).toHaveURL(new RegExp(`/${fixture}/$`));
        await expect(page.locator('.lp-dock-home')).toHaveAttribute('aria-current', 'page');
      }
    } finally { await context.close(); }
  });
}
