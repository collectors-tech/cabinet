import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';

const base = process.env.CABINET_BASE_URL || 'http://127.0.0.1:17881';
const ts = new Date().toISOString().replace(/[:.]/g,'-');
const outDir = path.join(process.cwd(), 'review', `cabinet-cycle-${ts}`);
fs.mkdirSync(outDir, { recursive: true });

const seed = ['/', '/sign-in', '/sign-up', '/forgot-password', '/terms', '/privacy', '/inventory', '/settings', '/dashboard'];
const screens = [];
const seen = new Set();

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
const evidence = [];
const failures = [];

const log = (x) => evidence.push({ time: new Date().toISOString(), ...x });

async function safeGoto(url) {
  for (let i = 0; i < 3; i++) {
    try {
      await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 15000 });
      await page.waitForTimeout(250);
      return;
    } catch {
      await page.waitForTimeout(200);
    }
  }
  await page.goto(url, { waitUntil: 'load', timeout: 20000 });
  await page.waitForTimeout(250);
}

function normPath(u) {
  try {
    const x = new URL(u, base);
    if (x.origin !== new URL(base).origin) return null;
    return x.pathname || '/';
  } catch { return null; }
}

async function discoverFromCurrent() {
  const links = await page.evaluate(() => {
    const vals = [];
    document.querySelectorAll('a[href]').forEach((a) => vals.push(a.getAttribute('href') || ''));
    return vals;
  });
  for (const h of links) {
    const p = normPath(h);
    if (!p || p.startsWith('/api/') || p.startsWith('/assets/')) continue;
    if (!seen.has(p)) { seen.add(p); screens.push(p); }
  }
}

async function listClickables() {
  return page.evaluate(() => {
    const canSee = (el) => { const r = el.getBoundingClientRect(); const s = getComputedStyle(el); return r.width>0&&r.height>0&&s.display!=='none'&&s.visibility!=='hidden'&&!el.hasAttribute('disabled')&&el.getAttribute('aria-disabled')!=='true'; };
    const els = Array.from(document.querySelectorAll('a,button,[role="button"],input[type="button"],input[type="submit"],summary,[tabindex],label[for]')).filter(canSee);
    return els.map((el, idx) => ({
      idx,
      tag: el.tagName.toLowerCase(),
      text: (el.innerText || el.getAttribute('aria-label') || el.getAttribute('title') || el.getAttribute('value') || '').trim().replace(/\s+/g,' ').slice(0,120),
      href: el.getAttribute('href') || '',
      role: el.getAttribute('role') || '',
      type: el.getAttribute('type') || ''
    }));
  });
}

async function clickByIndex(index) {
  return page.evaluate((index) => {
    const canSee = (el) => { const r = el.getBoundingClientRect(); const s = getComputedStyle(el); return r.width>0&&r.height>0&&s.display!=='none'&&s.visibility!=='hidden'&&!el.hasAttribute('disabled')&&el.getAttribute('aria-disabled')!=='true'; };
    const els = Array.from(document.querySelectorAll('a,button,[role="button"],input[type="button"],input[type="submit"],summary,[tabindex],label[for]')).filter(canSee);
    const el = els[index];
    if (!el) return { ok:false, error:'element-not-found' };
    try { el.click(); return { ok:true }; } catch (e) { return { ok:false, error:String(e.message||e) }; }
  }, index);
}

async function keyboardPath(screen) {
  await safeGoto(base + screen);
  const focusOrder = [];
  const activations = [];
  let prev = '';
  let repeat = 0;
  for (let i=0;i<30;i++) {
    await page.keyboard.press('Tab');
    const f = await page.evaluate(() => {
      const el = document.activeElement; if(!el) return null;
      return { tag: el.tagName.toLowerCase(), text:(el.innerText||el.getAttribute('aria-label')||el.getAttribute('title')||el.getAttribute('value')||'').trim().replace(/\s+/g,' ').slice(0,120) };
    });
    focusOrder.push(f);
    const k = f ? `${f.tag}|${f.text}` : 'null';
    if (k === prev) repeat++; else repeat = 0;
    prev = k;
    if (repeat >= 5) { activations.push({ focus:f, status:'pass', note:'tab-loop-breaker' }); break; }
    if (f && ['a','button','input','summary'].includes(f.tag)) {
      try {
        const beforeUrl = page.url();
        await page.keyboard.press('Enter');
        await page.waitForTimeout(100);
        const afterEnterUrl = page.url();
        await page.keyboard.press('Space');
        await page.waitForTimeout(100);
        const afterSpaceUrl = page.url();
        activations.push({ focus:f, beforeUrl, afterEnterUrl, afterSpaceUrl, status:'pass' });
      } catch (e) {
        const error = String(e.message||e);
        activations.push({ focus:f, status:'fail', error });
        failures.push({ screen, action:'keyboard activate', error, expected:'Focused interactive controls should activate without runtime errors.' });
      }
    }
  }
  for (let i=0;i<15;i++) await page.keyboard.press('Shift+Tab');
  log({ kind:'keyboard', screen, status:'pass', focusOrder, activations });
}

for (const p of seed) {
  if (!seen.has(p)) { seen.add(p); screens.push(p); }
}

for (let i=0; i<screens.length && i<20; i++) {
  const screen = screens[i];
  await safeGoto(base + screen);
  await discoverFromCurrent();
}

const audited = [...new Set(screens)].slice(0,20);
for (const screen of audited) {
  await safeGoto(base + screen);
  const slug = screen === '/' ? 'root' : screen.replace(/\//g,'_');
  await page.screenshot({ path: path.join(outDir, `screen-${slug}.png`), fullPage: true });
  const items = await listClickables();
  log({ kind:'inventory', screen, count: items.length, items });

  for (const item of items) {
    await safeGoto(base + screen);
    const beforeUrl = page.url();
    const result = await clickByIndex(item.idx);
    await page.waitForTimeout(120);
    const afterUrl = page.url();
    if (result.ok) {
      log({ kind:'click', screen, action:item, status:'pass', beforeUrl, afterUrl });
    } else {
      log({ kind:'click', screen, action:item, status:'fail', error:result.error, beforeUrl, afterUrl });
      failures.push({
        screen,
        action:`click ${item.tag} ${item.text}`,
        expected:'Clickable controls should execute without DOM/runtime click errors.',
        actual:result.error,
        errorText:result.error
      });
    }
  }

  await keyboardPath(screen);
}

const unique = [];
for (const f of failures) {
  const key = `${f.screen}|${f.action}|${f.actual || f.error}`;
  if (!unique.find(x => x.key === key)) unique.push({ key, ...f });
}

const report = {
  generatedAt: new Date().toISOString(),
  base,
  auditedScreens: audited,
  evidenceCount: evidence.length,
  passCount: evidence.filter(x => x.status !== 'fail').length,
  failureCount: unique.length,
  failures: unique,
  evidence
};

fs.writeFileSync(path.join(outDir, 'report.json'), JSON.stringify(report, null, 2));
const summary = [
  '# Cabinet UI Cycle Summary',
  '',
  `- Generated: ${report.generatedAt}`,
  `- Base: ${base}`,
  `- Audited screens (${audited.length}): ${audited.join(', ')}`,
  `- Evidence points: ${report.evidenceCount}`,
  `- Failures: ${report.failureCount}`,
  '',
  '## Failures',
  ...(unique.length ? unique.map((f) => `- ${f.screen} | ${f.action} | expected: ${f.expected || 'n/a'} | actual: ${f.actual || f.error} | error: ${f.errorText || f.actual || f.error}`) : ['- None'])
].join('\n');
fs.writeFileSync(path.join(outDir, 'SUMMARY.md'), summary);

await browser.close();
console.log(outDir);
