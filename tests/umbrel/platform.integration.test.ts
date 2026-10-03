// Copied into the pinned official Umbrel checkout by umbrel-platform.yml.
// Tests the actual app manager against the existing published store image.
import path from 'node:path'
import {beforeAll, afterAll, afterEach, expect, test} from 'vitest'
import fse from 'fs-extra'
import {$} from 'execa'
import pWaitFor from 'p-wait-for'
import yaml from 'js-yaml'
import createTestUmbreld from '../test-utilities/create-test-umbreld.js'
import runGitServer from '../test-utilities/run-git-server.js'

let platform: Awaited<ReturnType<typeof createTestUmbreld>> | undefined
let store: Awaited<ReturnType<typeof runGitServer>> | undefined
const appId = 'uqda-network'
let identity = ''
const base = 'http://127.0.0.1:8926'
let cookie = '', csrf = ''

async function ready() {
  await pWaitFor(async () => (await platform!.client.apps.state.query({appId})).state === 'ready',
    {interval:1000, timeout:180000})
  // App lifecycle readiness and gateway readiness are separate in Umbrel 2.0.
  await pWaitFor(async () => {
    try {
      const response = await request('/api/status', undefined, false)
      return response.status === 302 && response.headers.get('location')?.includes('/app-auth?') === true
    } catch { return false }
  }, {interval:1000, timeout:60000})
}
async function request(endpoint: string, value?: unknown, owner = true) {
  const jar = platform!.browserApi.defaults.options.cookieJar
  if (!jar) throw new Error('Official browser session cookie jar unavailable')
  const ownerCookie = owner ? await jar.getCookieString(base) : ''
  const response = await fetch(base + endpoint, {redirect:'manual',
    method:value === undefined ? 'GET' : 'POST', headers:{Cookie:[ownerCookie,cookie].filter(Boolean).join('; '), Origin:base,
      'Content-Type':'application/json', 'X-Uqda-CSRF':csrf},
    body:value === undefined ? undefined : JSON.stringify(value), signal:AbortSignal.timeout(35000)})
  return response
}
async function login() {
  const apps = await platform!.client.apps.list.query()
  const app = apps.find(app => app.id === appId)
  if (!app || 'error' in app) throw new Error('Installed app metadata unavailable')
  const response = await request('/api/login', {password:app.credentials.defaultPassword})
  expect(response.status).toBe(200)
  cookie = response.headers.get('set-cookie')!.split(';')[0]
  csrf = (await response.json()).csrf
}
async function nodeStatus() {
  let state: any
  await pWaitFor(async () => {
    const response = await request('/api/status')
    if (response.status !== 200) return false
    state = await response.json()
    return state.ready && state.tun?.enabled
  }, {interval:1000, timeout:60000})
  return state
}

beforeAll(async () => {
  store = await runGitServer()
  const source = process.env.UQDA_PACKAGE_SOURCE
  if (!source) throw new Error('UQDA_PACKAGE_SOURCE required')
  await fse.copy(path.join(source, 'uqda-network'), path.join(store.directory, appId))
  await fse.copy(path.join(source, 'umbrel-app-store.yml'), path.join(store.directory, 'umbrel-app-store.yml'))
  const git = $({cwd:store.directory})
  await git`git add .`
  await git`git commit -m ${'Add disposable Uqda store fixture'}`
  // Let the factory initialize valid API helper URLs before restarting on the
  // production port. autoStart:false constructs URLs with an undefined port.
  platform = await createTestUmbreld()
  // The official factory's port=0 intentionally skips LAN ingress. Use the
  // production internal port before starting so the real gateway is exercised.
  // The workflow has stopped the development service; all data remains temporary.
  await platform.instance.stop()
  platform.instance.port = 22080
  await platform.instance.start()
  await platform.signup()
  await platform.login()
  await platform.client.appStore.addRepository.mutate({url:store.url})
})
afterAll(async () => {
  await platform?.cleanup()
  await store?.close()
})

afterEach(async ({task}) => {
  if (task.result?.state !== 'fail') return
  // Keep bounded, credential-free startup diagnostics before temporary cleanup.
  console.error((await $`docker ps -a --format ${'{{.Names}}: {{.Status}}'}`).stdout)
  for (const service of ['core', 'dashboard']) {
    const result = await $({reject:false})`docker logs --tail 30 ${appId + '_' + service + '_1'}`
    console.error(result.stdout, result.stderr)
  }
})

test.sequential('install through actual Umbrel app manager and enforce both auth layers', async () => {
  await platform!.client.apps.install.mutate({appId})
  await ready()
  const app = (await platform!.client.apps.list.query()).find(app => app.id === appId)
  if (!app || 'error' in app) throw new Error('App installation failed')
  expect(app.appProxyAuth).toMatchObject({supported:true, defaultEnabled:true, override:null})
  const anonymous = await request('/api/status', undefined, false)
  expect(anonymous.status).toBe(302)
  expect(anonymous.headers.get('location')).toContain('/app-auth?')
  // A real owner login cookie authorizes Umbrel's gateway, but does not bypass
  // the independent dashboard password. Neither auth layer is disabled.
  expect((await request('/api/status')).status).toBe(401)
  await login()
  identity = (await nodeStatus()).identity.address
  expect(identity).toMatch(/^2[0-9a-f]*:/)
})

test.sequential('restart through Umbrel preserves identity and invalidates UI session', async () => {
  await platform!.client.apps.restart.mutate({appId})
  await ready()
  expect((await request('/api/status')).status).toBe(401)
  await login()
  expect((await nodeStatus()).identity.address).toBe(identity)
})

test.sequential('manifest update through Umbrel preserves identity', async () => {
  const manifestPath = path.join(store!.directory, appId, 'umbrel-app.yml')
  const manifest: any = yaml.load(await fse.readFile(manifestPath, 'utf8'))
  manifest.version = '26.0.4-umbrel.1-platform-test'
  await fse.writeFile(manifestPath, yaml.dump(manifest))
  const git = $({cwd:store!.directory})
  await git`git add .`
  await git`git commit -m ${'Change only disposable store manifest version'}`
  await platform!.instance.appStore.update()
  await platform!.client.apps.update.mutate({appId})
  await ready()
  await login()
  expect((await nodeStatus()).identity.address).toBe(identity)
  const installed = (await platform!.client.apps.list.query()).find(app => app.id === appId)
  expect(installed && !('error' in installed) && installed.version).toBe(manifest.version)
})

test.sequential('uninstall and fresh install produce a new identity after data removal', async () => {
  await platform!.client.apps.uninstall.mutate({appId})
  await pWaitFor(async () => !(await platform!.client.apps.list.query()).some(app => app.id === appId),
    {interval:1000, timeout:120000})
  const config = path.join(platform!.instance.dataDirectory, 'app-data', appId, 'data/config/uqda.conf')
  expect(await fse.pathExists(config)).toBe(false)
  cookie = ''; csrf = ''
  await platform!.client.apps.install.mutate({appId})
  await ready()
  await login()
  expect((await nodeStatus()).identity.address).not.toBe(identity)
})
