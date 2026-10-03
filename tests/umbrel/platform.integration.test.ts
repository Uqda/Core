// Copied into the pinned official Umbrel checkout by umbrel-platform.yml.
// Tests the actual app manager against the published image and rebuilt source.
import path from 'node:path'
import net from 'node:net'
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
// Preserve the real previous-release upgrade boundary even when packaging moves
// to the next wrapper version. This immutable image is used only in CI fixtures.
const publishedImage = 'ghcr.io/uqda/core:26.0.4-umbrel.1@sha256:e68a42d1f2063e408272f96189a45c2f68beb09fee659700693b4ab6694a78ad'
let packageVersion = ''
let currentPublishedImage = ''
const base = 'http://127.0.0.1:8926'
let cookie = '', csrf = ''
let serviceEntry: any

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
async function request(endpoint: string, value?: unknown, owner = true, proof = csrf) {
  const jar = platform!.browserApi.defaults.options.cookieJar
  if (!jar) throw new Error('Official browser session cookie jar unavailable')
  const ownerCookie = owner ? await jar.getCookieString(base) : ''
  const response = await fetch(base + endpoint, {redirect:'manual',
    method:value === undefined ? 'GET' : 'POST', headers:{Cookie:[ownerCookie,cookie].filter(Boolean).join('; '), Origin:base,
      'Content-Type':'application/json', 'X-Uqda-CSRF':proof},
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

const peerContainer = 'uqda-disposable-remote-peer'
async function peerJson(program: string, input: unknown) {
  // Request cookies/proofs travel only over stdin, never command arguments/logs.
  const result = await $({input:JSON.stringify(input)})`docker exec -i ${peerContainer} python3 -c ${program}`
  return JSON.parse(result.stdout)
}
const peerControl = (request: unknown) => peerJson(
  "import json,sys; from control import socket_json; print(json.dumps(socket_json('/run/uqda-control/control.sock', json.load(sys.stdin), timeout=32)))",
  request,
)
const peerHttp = (request: unknown) => peerJson(`
import http.client, json, sys
request = json.load(sys.stdin)
try:
    client = http.client.HTTPConnection(request['address'], request['port'], timeout=5)
    client.request('GET', request['path'], headers={'Cookie':request.get('cookie', ''), 'X-Uqda-CSRF':request.get('proof', '')})
    response = client.getresponse()
    body = response.read(8192).decode('utf-8', errors='replace')
    print(json.dumps({'status':response.status, 'location':response.getheader('location'), 'body':body}))
    client.close()
except OSError:
    print(json.dumps({'unreachable':True}))
`, request)

async function unusedHostPort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const reservation = net.createServer()
    reservation.once('error', reject)
    reservation.listen(0, '0.0.0.0', () => {
      const address = reservation.address()
      if (!address || typeof address === 'string') return reject(new Error('Port reservation failed'))
      reservation.close(error => error ? reject(error) : resolve(address.port))
    })
  })
}

beforeAll(async () => {
  store = await runGitServer()
  const source = process.env.UQDA_PACKAGE_SOURCE
  if (!source) throw new Error('UQDA_PACKAGE_SOURCE required')
  await fse.copy(path.join(source, 'uqda-network'), path.join(store.directory, appId))
  await fse.copy(path.join(source, 'umbrel-app-store.yml'), path.join(store.directory, 'umbrel-app-store.yml'))
  const manifest: any = yaml.load(await fse.readFile(path.join(store.directory, appId, 'umbrel-app.yml'), 'utf8'))
  packageVersion = manifest.version
  const composePath = path.join(store.directory, appId, 'docker-compose.yml')
  const compose: any = yaml.load(await fse.readFile(composePath, 'utf8'))
  currentPublishedImage = compose.services.core.image
  const sourceImage = process.env.UQDA_PLATFORM_IMAGE
  if (sourceImage) {
    if (!/^127\.0\.0\.1:5000\/uqda-validation@sha256:[a-f0-9]{64}$/.test(sourceImage)) {
      throw new Error('Expected a digest-pinned disposable local image')
    }
    for (const service of ['core', 'dashboard']) compose.services[service].image = sourceImage
    await fse.writeFile(composePath, yaml.dump(compose))
  }
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
  if (process.env.UQDA_PLATFORM_IMAGE) {
    // Fail locally first if the official UI build is missing. A healthy API
    // fixture alone cannot prove that the actual Umbrel dashboard is served.
    const home = await fetch('http://127.0.0.1/', {signal:AbortSignal.timeout(5000)})
    expect(home.status).toBe(200)
    expect((await home.text()).toLowerCase()).toContain('<html')
  }
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
  if (process.env.UQDA_PLATFORM_IMAGE || packageVersion !== '26.0.4-umbrel.1') {
    // Co-hosted apps can receive cookies because cookies ignore TCP ports.
    // Even a real owner cookie plus the UI cookie cannot recover its proof.
    const session = await request('/api/session', undefined, true, '')
    expect(session.status).toBe(200)
    expect(await session.json()).toMatchObject({authenticated:false, csrf:''})
    expect((await request('/api/status', undefined, true, '')).status).toBe(401)
    expect((await nodeStatus()).identity.address).toBe(identity)
  }
  if (process.env.UQDA_PLATFORM_IMAGE) {
    const state = await nodeStatus()
    const response = await request('/api/services/add', {
      revision:state.services.revision, name:'My files', kind:'https', port:8443,
    })
    expect(response.status).toBe(200)
    serviceEntry = (await response.json()).services.items[0]
    expect(serviceEntry.endpoint).toBe(`https://[${identity}]:8443/`)
    expect((await request('/api/services/probe', {id:serviceEntry.id}, true, '')).status).toBe(403)
    expect((await request('/api/services/remove', {revision:state.services.revision,id:serviceEntry.id}, true, '')).status).toBe(403)
    expect((await nodeStatus()).settings.revision).toBe(state.settings.revision)
  }
})

test.sequential('restart through Umbrel preserves identity and invalidates UI session', async () => {
  await platform!.client.apps.restart.mutate({appId})
  await ready()
  expect((await request('/api/status')).status).toBe(401)
  await login()
  expect((await nodeStatus()).identity.address).toBe(identity)
  if (serviceEntry) expect((await nodeStatus()).services.items).toContainEqual(serviceEntry)
})

test.sequential('actual Umbrel dashboard and authenticated app gateway work over a private overlay', async () => {
  const image = process.env.UQDA_PLATFORM_IMAGE
  // Published wrapper .2 lacks the new source workflow; it remains unchanged.
  if (!image) return
  const before = await nodeStatus()
  const port = await unusedHostPort()
  const group = 'disposable-umbrel-remote-test-group-12345'
  const peers = [`tls://host.docker.internal:${port}`]
  let created = false
  try {
    const applied = await request('/api/settings', {revision:before.settings.revision,
      peers:before.settings.peers,listen:[`tls://0.0.0.0:${port}`],mode:'private',group_password:group})
    expect(applied.status).toBe(200)
    await $`docker run --detach --name ${peerContainer} --cap-add NET_ADMIN --device /dev/net/tun --add-host host.docker.internal:host-gateway --tmpfs /etc/uqda:mode=0700 --tmpfs /run/uqda-core:mode=0700 --tmpfs /run/uqda-control:mode=0700 ${image} control.py`
    created = true
    await pWaitFor(async () => {
      try { return (await peerControl({action:'status'})).result?.ready === true } catch { return false }
    }, {interval:1000,timeout:60000})
    const configurePeer = async (password: string) => {
      const state = await peerControl({action:'status'})
      const result = await peerControl({action:'apply',settings:{revision:state.result.settings.revision,
        peers,listen:[],mode:'private',group_password:password}})
      expect(result.ok).toBe(true)
    }
    await configurePeer(group)
    const dashboardRequest = {address:identity,port:80,path:'/'}
    let lastDashboard: any
    try {
      await pWaitFor(async () => {
        lastDashboard = await peerHttp(dashboardRequest)
        return lastDashboard.status === 200
      }, {interval:1000,timeout:45000})
    } catch {
      // Only credential-free response metadata, never cookies/config or body.
      throw new Error(`Remote dashboard failed: ${JSON.stringify({status:lastDashboard?.status,
        unreachable:lastDashboard?.unreachable})}`)
    }
    const home = await peerHttp(dashboardRequest)
    expect(home.body.toLowerCase()).toContain('<html')
    const anonymous = await peerHttp({address:identity,port:8926,path:'/api/status'})
    expect(anonymous.status).toBe(302)
    expect(anonymous.location.startsWith(`http://[${identity}]:2000/app-auth?`)).toBe(true)
    const jar = platform!.browserApi.defaults.options.cookieJar!
    const ownerCookie = await jar.getCookieString(base)
    const owner = {address:identity,port:8926,path:'/api/status',cookie:ownerCookie}
    expect((await peerHttp(owner)).status).toBe(401)
    const authenticated = {...owner,cookie:[ownerCookie,cookie].join('; '),proof:csrf}
    const result = await peerHttp(authenticated)
    expect(result.status).toBe(200)
    expect(JSON.parse(result.body).identity.address).toBe(identity)
    expect((await peerHttp({...authenticated,proof:''})).status).toBe(401)
    const local = await request('/api/umbrel/probe', {})
    expect(local.status).toBe(200)
    expect(await local.json()).toMatchObject({scope:'local',remote_verified:false,
      ports:[{port:80,tcp_reachable:true},{port:443,tcp_reachable:true},{port:2000,tcp_reachable:true}]})
    await configurePeer('wrong-umbrel-remote-test-group-12345')
    expect((await peerHttp(dashboardRequest)).unreachable).toBe(true)
    await configurePeer(group)
    await pWaitFor(async () => (await peerHttp(dashboardRequest)).status === 200, {interval:1000,timeout:45000})
    expect((await nodeStatus()).identity.address).toBe(identity)
  } finally {
    if (created) await $`docker rm -f ${peerContainer}`
    const state = await nodeStatus()
    expect((await request('/api/settings', {revision:state.settings.revision,
      peers:before.settings.peers,listen:before.settings.listen,mode:'private'})).status).toBe(200)
  }
}, 180_000)

test.sequential('manifest update and previous-to-current published image upgrade preserve identity', async () => {
  const manifestPath = path.join(store!.directory, appId, 'umbrel-app.yml')
  const manifest: any = yaml.load(await fse.readFile(manifestPath, 'utf8'))
  const sourceImage = process.env.UQDA_PLATFORM_IMAGE
  // The source matrix initially installs a local rebuild. Now move to the real
  // previous release, then upgrade to the actual new public digest, not a local
  // stand-in. This covers precisely the update store users will receive.
  const images = sourceImage ? [publishedImage, currentPublishedImage] : ['']
  for (const [index, image] of images.entries()) {
    manifest.version = `${packageVersion}-platform-test-${index}`
    await fse.writeFile(manifestPath, yaml.dump(manifest))
    if (image) {
      const composePath = path.join(store!.directory, appId, 'docker-compose.yml')
      const compose: any = yaml.load(await fse.readFile(composePath, 'utf8'))
      for (const service of ['core', 'dashboard']) compose.services[service].image = image
      await fse.writeFile(composePath, yaml.dump(compose))
    }
    const git = $({cwd:store!.directory})
    await git`git add .`
    await git`git commit -m ${'Update disposable store fixture stage ' + index}`
    await platform!.instance.appStore.update()
    await platform!.client.apps.update.mutate({appId})
    await ready()
    if (image) {
      for (const service of ['core', 'dashboard']) {
        const actual = await $`docker inspect --format ${'{{.Config.Image}}'} ${appId + '_' + service + '_1'}`
        expect(actual.stdout.trim()).toBe(image)
      }
    }
    await login()
    expect((await nodeStatus()).identity.address).toBe(identity)
    if (serviceEntry) {
      const saved = path.join(platform!.instance.dataDirectory, 'app-data', appId, 'data/config/services.json')
      expect(await fse.readJson(saved)).toContainEqual({id:serviceEntry.id,name:'My files',kind:'https',port:8443})
    }
    const installed = (await platform!.client.apps.list.query()).find(app => app.id === appId)
    expect(installed && !('error' in installed) && installed.version).toBe(manifest.version)
    if (sourceImage && image !== publishedImage) {
      expect((await request('/api/status', undefined, true, '')).status).toBe(401)
    }
  }
})

test.sequential('uninstall and fresh install produce a new identity after data removal', async () => {
  await platform!.client.apps.uninstall.mutate({appId})
  await pWaitFor(async () => !(await platform!.client.apps.list.query()).some(app => app.id === appId),
    {interval:1000, timeout:120000})
  const config = path.join(platform!.instance.dataDirectory, 'app-data', appId, 'data/config/uqda.conf')
  expect(await fse.pathExists(config)).toBe(false)
  expect(await fse.pathExists(path.join(path.dirname(config), 'services.json'))).toBe(false)
  cookie = ''; csrf = ''
  await platform!.client.apps.install.mutate({appId})
  await ready()
  await login()
  expect((await nodeStatus()).identity.address).not.toBe(identity)
})
