import { expect, test } from '@playwright/test';

const themes = ['classic', 'aurora', 'paper', 'quiet', 'terminal'];
const styles = ['base', 'underlined', 'box'];

// Compare rendered pixels, not just the presence of a CSS shadow. An outer
// underline can exist in computed styles while its scroll container clips it.
async function underlinePixels(page, link) {
  const before = await link.screenshot();
  const suppress = await page.addStyleTag({ content: `
    .lp-nav .lp-nav-link--active { box-shadow: none !important; }
    .lp-nav .lp-nav-link--active::after { background: none !important; }
  ` });
  const after = await link.screenshot();
  await suppress.evaluate(el => el.remove());
  return page.evaluate(async ([a, b]) => {
    async function decode(base64) {
      const image = new Image();
      image.src = 'data:image/png;base64,' + base64;
      await image.decode();
      const canvas = document.createElement('canvas');
      canvas.width = image.width; canvas.height = image.height;
      const ctx = canvas.getContext('2d');
      ctx.drawImage(image, 0, 0);
      return { width: image.width, height: image.height, data: ctx.getImageData(0, 0, image.width, image.height).data };
    }
    const on = await decode(a), off = await decode(b);
    let changed = 0, outsideBottom = 0;
    const columns = new Set();
    for (let y = 0; y < on.height; y++) for (let x = 0; x < on.width; x++) {
      const i = (y * on.width + x) * 4;
      if ([0, 1, 2].some(c => Math.abs(on.data[i + c] - off.data[i + c]) > 8)) {
        changed++; columns.add(x);
        if (y < on.height - 6) outsideBottom++;
      }
    }
    return { changed, outsideBottom, columns: columns.size, width: on.width };
  }, [before.toString('base64'), after.toString('base64')]);
}

for (const everything of [false, true]) for (const theme of themes) for (const activeStyle of styles) for (const colorScheme of ['light', 'dark']) {
  test(`${theme} ${everything ? "Everything" : "legacy"} bottom ${activeStyle} renders in ${colorScheme}`, async ({ page }) => {
    await page.emulateMedia({ colorScheme, reducedMotion: 'reduce' });
    for (const width of [1440, 390]) {
      await page.setViewportSize({ width, height: 844 });
      await page.goto(`/${theme}${everything ? "-everything" : ""}-bottom-${activeStyle}/notes/components/`);
      await page.evaluate(() => document.fonts.ready);
      const inline = page.locator('.lp-nav-links > .lp-nav-link--active').first();
      const inMenu = everything && !(await inline.isVisible());
      if (inMenu) await page.locator('.lp-section-picker summary').click();
      const link = inMenu ? page.locator('.lp-section-menu > .lp-nav-link--active').first() : inline;
      await expect(link).toBeVisible();
      await expect(link).toHaveClass(new RegExp(`lp-nav-active-${activeStyle}`));
      const state = await link.evaluate(el => {
        const s = getComputedStyle(el), row = el.parentElement.getBoundingClientRect(), r = el.getBoundingClientRect();
        const probe = document.createElement('span');
        probe.style.color = 'var(--lp-accent)'; document.body.append(probe);
        const accent = getComputedStyle(probe).color; probe.remove();
        return { color: s.color, accent, weight: Number(s.fontWeight), background: s.backgroundColor, image: s.backgroundImage,
          contained: r.left >= row.left - 1 && r.right <= row.right + 1,
          inactiveWeight: Number(getComputedStyle(el.parentElement.querySelector('a:not(.lp-nav-link--active)')).fontWeight) };
      });
      expect(state.contained).toBe(true);
      if (activeStyle === 'underlined') {
        const pixels = await underlinePixels(page, link);
        expect(pixels.changed).toBeGreaterThan(8);
        expect(pixels.columns).toBeGreaterThan(pixels.width / 4);
        expect(pixels.outsideBottom).toBe(0);
      } else if (activeStyle === 'base') {
        expect(state.color).toBe(state.accent);
        if (theme === 'quiet') expect(state.weight).toBeGreaterThan(state.inactiveWeight);
      } else {
        expect(state.background !== 'rgba(0, 0, 0, 0)' || state.image !== 'none').toBe(true);
      }
    }
  });
}
