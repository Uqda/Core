// Exercise the actual dashboard and daemon started by docker_smoke.py.
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const fs = require('node:fs');
const path = require('node:path');

(async () => {
  const browser = await chromium.launch({headless:true});
  try {
    const page = await browser.newPage({viewport:{width:1280,height:1120}});
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto('http://127.0.0.1:8926');
    await page.fill('#loginPassword', process.env.UQDA_TEST_PASSWORD);
    await page.click('#loginForm button');
    await page.waitForFunction(() => document.querySelector('#address').textContent !== '—');
    const before = await page.locator('#address').textContent();
    const output = process.env.UQDA_SCREENSHOT_DIR;
    if (output) {
      fs.mkdirSync(output, {recursive:true});
      await page.screenshot({path:path.join(output,'dashboard-en.png'),fullPage:true});
    }
    await page.click('#language');
    if (await page.locator('html').getAttribute('dir') !== 'rtl') throw Error('RTL not active');
    if (output) await page.screenshot({path:path.join(output,'dashboard-ar.png'),fullPage:true});
    await page.fill('#groupPassword','browser-private-group-secret-1234');
    await page.click('#save');
    await page.waitForFunction(() => document.querySelector('#notice').textContent.includes('تم حفظ'));
    if (await page.locator('#address').textContent() !== before) throw Error('Identity changed');
    await page.setViewportSize({width:390,height:844});
    if (output) await page.screenshot({path:path.join(output,'dashboard-mobile.png'),fullPage:true});
    if (await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)) throw Error('Mobile overflow');
    await page.click('#logout');
    await page.waitForFunction(() => !document.querySelector('#login').hidden);
    if (errors.length) throw Error(errors.join('; '));
    console.log('PASS: browser login, real status, Arabic RTL, settings save, stable identity, mobile width, logout');
  } finally {
    await browser.close();
  }
})().catch(error => {console.error(error);process.exit(1);});
