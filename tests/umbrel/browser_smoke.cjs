// Exercise the actual dashboard and daemon started by docker_smoke.py.
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');

async function until(check) {
  const deadline = Date.now() + 30000;
  while (Date.now() < deadline) {
    if (await check()) return;
    await new Promise(resolve => setTimeout(resolve, 150));
  }
  throw Error('Browser state did not become ready');
}

(async () => {
  const browser = await chromium.launch({headless:true});
  try {
    const page = await browser.newPage({viewport:{width:1280,height:1120}});
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto('http://127.0.0.1:8926');
    await page.fill('#loginPassword', process.env.UQDA_TEST_PASSWORD);
    await page.click('#loginForm button');
    await until(async () => await page.locator('#address').textContent() !== '—');
    const before = await page.locator('#address').textContent();
    if (await page.locator('#umbrelHttpAddress').textContent() !== 'http://[' + before + ']/') throw Error('Wrong remote Umbrel address');
    await page.click('#checkUmbrel');
    await until(async () => await page.locator('#umbrelChecks li').count() === 3);
    await page.fill('#serviceName', 'My files <not HTML>');
    await page.selectOption('#serviceKind', 'https');
    await page.fill('#servicePort', '8443');
    await page.click('#addService');
    await until(async () => await page.locator('.service-entry').count() === 1);
    if (!(await page.locator('.service-entry h3').textContent()).includes('<not HTML>')) throw Error('Service name was not rendered as text');
    if (!(await page.locator('.service-entry code').first().textContent()).includes('https://[' + before + ']:8443/')) throw Error('Wrong IPv6 service address');
    if (!(await page.locator('.service-entry').textContent()).includes('Remote access: not verified')) throw Error('Remote access incorrectly claimed');
    await page.click('.service-entry button[data-probe="true"]');
    await until(async () => !(await page.locator('.service-entry').textContent()).includes('Local TCP: not checked'));
    // Cookies ignore ports. Exercise an actual co-hosted untrusted HTTP origin,
    // without printing its captured test cookie or the origin-scoped proof.
    let capturedCookie = '';
    const otherApp = http.createServer((request,response) => {
      capturedCookie = request.headers.cookie || '';
      response.end('<!doctype html><title>Disposable other-app fixture</title>');
    });
    await new Promise(resolve => otherApp.listen(0,'127.0.0.1',resolve));
    try {
      await page.goto('http://127.0.0.1:' + otherApp.address().port);
      if (!capturedCookie.includes('uqda_session=')) throw Error('Co-host cookie fixture did not capture its test session');
      if (await page.evaluate(() => sessionStorage.getItem('uqda-session-proof')) !== null) throw Error('Session proof leaked across origins');
      const bootstrap = await fetch('http://127.0.0.1:8926/api/session', {headers:{Cookie:capturedCookie}});
      const session = await bootstrap.json();
      if (session.authenticated || session.csrf) throw Error('Cookie-only session disclosed authentication proof');
      if ((await fetch('http://127.0.0.1:8926/api/status', {headers:{Cookie:capturedCookie}})).status !== 401) throw Error('Cookie-only read allowed');
      await page.goto('http://127.0.0.1:8926');
      await until(async () => await page.locator('#address').textContent() === before);
      await page.reload();
      await until(async () => await page.locator('#address').textContent() === before);
    } finally { await new Promise(resolve => otherApp.close(resolve)); }
    const output = process.env.UQDA_SCREENSHOT_DIR;
    if (output) {
      fs.mkdirSync(output, {recursive:true});
      await page.screenshot({path:path.join(output,'dashboard-en.png'),fullPage:true});
    }
    await page.click('#language');
    if (await page.locator('html').getAttribute('dir') !== 'rtl') throw Error('RTL not active');
    if (!(await page.locator('.service-entry').textContent()).includes('الوصول عن بُعد: لم يُتحقق منه')) throw Error('Arabic service status missing');
    if (output) await page.screenshot({path:path.join(output,'dashboard-ar.png'),fullPage:true});
    await page.fill('#groupPassword','browser-private-group-secret-1234');
    await page.click('#save');
    await until(async () => (await page.locator('#notice').textContent()).includes('تم حفظ'));
    if (await page.locator('#address').textContent() !== before) throw Error('Identity changed');
    for (const id of ['peers','listeners','networkMode','groupPassword','save']) {
      if (!await page.locator('#' + id).isEnabled()) throw Error('Settings remained locked after save: ' + id);
    }
    await page.fill('#groupPassword','unsaved-secret-must-clear');
    await page.setViewportSize({width:390,height:844});
    if (output) await page.screenshot({path:path.join(output,'dashboard-mobile.png'),fullPage:true});
    if (await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)) throw Error('Mobile overflow');
    await page.click('#logout');
    await page.locator('#login').waitFor({state:'visible'});
    if (await page.locator('#groupPassword').inputValue()) throw Error('Logout retained a secret');
    if (await page.locator('#address').textContent() !== '—') throw Error('Logout retained identity');
    if (errors.length) throw Error(errors.join('; '));
    console.log('PASS: browser login, real status, Arabic RTL, settings save, stable identity, mobile width, logout');
  } finally {
    await browser.close();
  }
})().catch(error => {console.error(error);process.exit(1);});
