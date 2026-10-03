// Dependency-free state regression tests. Real DOM/layout is covered by browser_smoke.cjs.
'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.resolve(__dirname, '../..');
const source = fs.readFileSync(path.join(root, 'contrib/umbrel/web/app.js'), 'utf8');
const html = fs.readFileSync(path.join(root, 'contrib/umbrel/web/index.html'), 'utf8');

function element() {
  return {hidden:false, disabled:false, checked:false, value:'', textContent:'', dataset:{},
    children:[], events:{}, append(...values) { this.children.push(...values); },
    replaceChildren(...values) { this.children = values; }, querySelectorAll() { return []; }, reset() {},
    addEventListener(name, handler) { this.events[name] = handler; }};
}
const ids = Object.fromEntries([...html.matchAll(/id="([^"]+)"/g)].map(match => [match[1], element()]));
const fields = ['peers', 'listeners', 'networkMode', 'groupPassword', 'confirmPublic', 'save'].map(id => ids[id]);
ids.settingsForm.querySelectorAll = () => fields;
ids.settingsForm.reset = () => {
  for (const field of fields) { field.value = ''; field.checked = false; }
  ids.networkMode.value = 'private';
};
ids.loginForm.reset = () => { ids.loginPassword.value = ''; };
const reply = (data, status = 200) => ({ok:status === 200, status, json:async () => data});
const sessionValues = new Map();
const context = vm.createContext({document:{documentElement:{}, getElementById:id => ids[id],
  querySelectorAll:() => [], createElement:element, createTextNode:text => text},
  localStorage:{getItem:() => 'en', setItem:() => {}}, navigator:{language:'en'},
  sessionStorage:{getItem:key => sessionValues.get(key) || null, setItem:(key,value) => sessionValues.set(key,value), removeItem:key => sessionValues.delete(key)},
  setInterval:() => {}, confirm:() => true, fetch:async () => reply({authenticated:false, csrf:''})});
const run = code => vm.runInContext(code, context);
const state = {ready:true, identity:{address:'200::123',build_version:'26.0.4'}, tun:{enabled:true,name:'uqda0'},
  umbrel_access:{enabled:true,reason:'ready',https_url:'https://[200::123]/',http_url:'http://[200::123]/',network_command:'sudo uqda test 200::123'},
  peers:[], services:{revision:'s1',items:[]}, settings:{private:true,editable:true,revision:'r1',peers:[],listen:[]}};
context.state = state;
const flush = () => new Promise(resolve => setImmediate(resolve));

(async () => {
  vm.runInContext(source, context);
  await flush();
  run('csrf = "tab-local-proof"; storeSessionProof(csrf)');
  context.fetch = async (_path, options) => {
    assert.equal(options.headers['X-Uqda-CSRF'], 'tab-local-proof', 'GET must require origin-scoped session proof');
    return reply({authenticated:true,csrf:'tab-local-proof'});
  };
  await run('api("/api/session")');
  run('authenticated(true); render(state)');
  assert.equal(ids.umbrelHttpAddress.textContent, 'http://[200::123]/');
  assert.equal(ids.checkUmbrel.disabled, false);
  context.publicState = {...state, settings:{...state.settings,private:false,revision:'public'},umbrel_access:{enabled:false,reason:'private_required'}};
  run('render(publicState)');
  assert.equal(ids.umbrelAddresses.hidden, true);
  assert.equal(ids.umbrelHttpAddress.textContent, '');
  assert.equal(ids.checkUmbrel.disabled, true);
  run('render(state)');
  const service = {id:'0123456789abcdef',name:'<img onerror=alert(1)>',kind:'https',port:8443,endpoint:'https://[200::123]:8443/',network_command:'sudo uqda test 200::123',service_command:"curl --head 'https://[200::123]:8443/'"};
  context.withService = {...state,services:{revision:'s2',items:[service]}};
  run('render(withService)');
  assert.equal(ids.serviceList.children[0].children[0].textContent, service.name + ' · HTTPS · 8443');
  context.fetch = async () => reply({id:service.id,tcp_reachable:true,scope:'local',remote_verified:false});
  await run('serviceRequest("/api/services/probe", {id:"0123456789abcdef"}, true)');
  assert.equal(run('serviceChecks.get("0123456789abcdef")'), true);
  assert.equal(ids.serviceList.children[0].children[7].textContent, 'Remote access: not verified — test from your other Uqda device.');
  run('render(state)');
  assert.equal(run('serviceChecks.size'), 0, 'changed service list must invalidate local check');
  ids.groupPassword.value = 'secret-entered-by-user';
  context.fetch = async () => reply(state);
  await ids.settingsForm.events.submit({preventDefault() {}});
  assert(fields.every(field => !field.disabled), 'all settings must unlock immediately after save');
  assert.equal(ids.groupPassword.value, '', 'saved secret must be cleared');

  let completeRestart;
  context.fetch = () => new Promise(resolve => { completeRestart = resolve; });
  const restarting = ids.restart.events.click();
  assert(fields.every(field => field.disabled), 'restart must lock all settings');
  assert(ids.logout.disabled, 'logout must not race a configuration mutation');
  completeRestart(reply(state));
  await restarting;
  assert(fields.every(field => !field.disabled), 'restart must unlock all settings immediately');

  context.fetch = async () => reply({error:'Settings rejected'}, 400);
  ids.groupPassword.value = 'retryable-secret';
  await ids.settingsForm.events.submit({preventDefault() {}});
  assert(fields.every(field => !field.disabled), 'failed save must unlock settings');
  assert.equal(ids.groupPassword.value, 'retryable-secret', 'validation failure must keep the draft');

  let completePoll;
  context.fetch = () => new Promise(resolve => { completePoll = resolve; });
  const polling = run('refresh()');
  run('authenticated(false)');
  completePoll(reply(state));
  await polling;
  assert.equal(ids.address.textContent, '—', 'late status must not restore logged-out identity');
  assert.equal(ids.groupPassword.value, '', 'logout must clear unsaved secrets');
  assert.equal(run('snapshot'), null);
  assert.equal(ids.umbrelHttpAddress.textContent, '', 'logout must clear remote Umbrel address');
  assert.equal(run('csrf'), '');
  assert.equal(sessionValues.size, 0, 'logout must clear tab-local proof');

  run('authenticated(true); render(state)');
  context.fetch = async () => reply({error:'Sign in'}, 401);
  ids.groupPassword.value = 'expired-session-secret';
  await ids.settingsForm.events.submit({preventDefault() {}});
  assert(ids.dashboard.hidden, 'expired sessions must hide the dashboard');
  assert.equal(ids.groupPassword.value, '', 'expired sessions must clear secrets');
  assert(fields.every(field => field.disabled), 'no settings access without a new snapshot');

  assert.match(html, /data-i18n="purposeTitle"/);
  for (const language of ['en', 'ar']) {
    const keys = new Set([...html.matchAll(/data-i18n="([^"]+)"/g)].map(match => match[1]));
    for (const key of keys) {
      assert(run(`text.${language}.${key}.length > 0`), `missing ${language} guide translation: ${key}`);
    }
  }
  console.log('PASS: save/restart control locking, failed save, logout/expiry secret cleanup, stale polling, bilingual guide');
})().catch(error => { console.error(error); process.exitCode = 1; });
