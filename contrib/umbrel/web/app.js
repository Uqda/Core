'use strict';
const text = {
  en: {eyebrow:'YOUR UMBREL. CONNECTED.',heading:'A connection you control.',intro:'Manage your encrypted IPv6 connection from one place.',signOut:'Sign out',loginTitle:'Open your dashboard',loginHint:"Use the app password shown in Umbrel’s Uqda app details.",appPassword:'App password',signIn:'Sign in',nodeStatus:'Your node',nodeAddress:'Uqda IPv6 address',copy:'Copy',connectedPeers:'Connected peers',interface:'Network interface',coreVersion:'Core version',mode:'Access mode',accessHint:'A connected transport peer does not prove private-group connectivity. Services also need to listen on IPv6 and allow access through their firewall.',settings:'Connection settings',savedOnDevice:'Saved on this device',firstRun:'New nodes start isolated. Add a trusted peer and your shared group password to connect.',peerAddresses:'Peer addresses',peersHint:'One address per line. TLS, TCP, QUIC, WS and WSS are supported. No peer is added automatically.',privateGroup:'Private group',publicNetwork:'Public network',groupPassword:'Shared group password',groupHint:'Use the same strong secret on all members. Leave blank to keep the current secret. Stored secrets are never shown here.',publicConfirm:'I understand that public mode removes private-group protection. This can make services listening on my Uqda IPv6 address reachable by other public nodes.',advanced:'Incoming connections',listenerAddresses:'Listener addresses (optional)',listenerHint:'Leave empty for outbound-only connections. Use TLS, TCP or QUIC and a port from 1024 to 65535. Router/firewall changes are separate.',save:'Save and reconnect',restartHint:'Saving restarts Core. Your node identity stays the same.',peers:'Peers',refresh:'Refresh',health:'Local health',restart:'Restart Core',project:'Project & documentation ↗',connected:'Connected',isolated:'No connected peers',starting:'Starting',noPeers:'No peers yet. Add a trusted peer in connection settings.',daemon:'Core responding',identity:'Node identity available',tun:'TUN interface enabled',peerHealth:'At least one transport peer connected',up:'Connected',down:'Disconnected',inbound:'Incoming',outbound:'Outgoing',saved:'Settings saved. Core reconnected with the same identity.',restarted:'Core restarted.',busy:'Applying changes…',copied:'Address copied.',copyFailed:'Select the address and copy it manually.',advancedConfig:'Advanced peer settings detected. Browser editing is disabled to preserve the original configuration.',restartConfirm:'Restart Core? Connections will briefly disconnect.',networkError:'The dashboard is unavailable. Try again shortly.'},
  ar: {eyebrow:'جهاز UMBREL متصل بشبكتك',heading:'اتصال بإدارتك.',intro:'أدر اتصال IPv6 المشفّر من مكان واحد.',signOut:'تسجيل الخروج',loginTitle:'افتح لوحة الإدارة',loginHint:'استخدم كلمة مرور التطبيق الظاهرة في تفاصيل Uqda داخل Umbrel.',appPassword:'كلمة مرور التطبيق',signIn:'تسجيل الدخول',nodeStatus:'عُقدتك',nodeAddress:'عنوان Uqda IPv6',copy:'نسخ',connectedPeers:'العُقد المتصلة',interface:'واجهة الشبكة',coreVersion:'إصدار Core',mode:'وضع الوصول',accessHint:'اتصال النقل بعقدة أخرى لا يثبت اتصال المجموعة الخاصة. الخدمات تحتاج أيضًا للاستماع على IPv6 والسماح بالوصول عبر الجدار الناري.',settings:'إعدادات الاتصال',savedOnDevice:'محفوظة على هذا الجهاز',firstRun:'العُقد الجديدة تبدأ معزولة. أضف عقدة موثوقة وكلمة مرور مجموعتك المشتركة للاتصال.',peerAddresses:'عناوين العُقد',peersHint:'عنوان واحد في كل سطر. يدعم TLS وTCP وQUIC وWS وWSS. لا تُضاف أي عقدة تلقائيًا.',privateGroup:'مجموعة خاصة',publicNetwork:'شبكة عامة',groupPassword:'كلمة مرور المجموعة المشتركة',groupHint:'استخدم سرًا قويًا ومطابقًا لدى جميع الأعضاء. اترك الحقل فارغًا للحفاظ على السر الحالي. الأسرار المحفوظة لا تظهر هنا.',publicConfirm:'أفهم أن الوضع العام يزيل حماية المجموعة الخاصة، وقد يتيح للعُقد العامة الوصول إلى الخدمات التي تستمع على عنوان Uqda IPv6 الخاص بجهازي.',advanced:'الاتصالات الواردة',listenerAddresses:'عناوين الاستماع (اختياري)',listenerHint:'اتركه فارغًا للاتصالات الصادرة فقط. استخدم TLS أو TCP أو QUIC ومنفذًا بين 1024 و65535. إعداد الراوتر والجدار الناري منفصل.',save:'حفظ وإعادة الاتصال',restartHint:'الحفظ يعيد تشغيل Core ويحافظ على هوية العُقدة.',peers:'العُقد',refresh:'تحديث',health:'حالة الجهاز',restart:'إعادة تشغيل Core',project:'المشروع والتوثيق ↗',connected:'متصل',isolated:'لا توجد عُقد متصلة',starting:'جارٍ التشغيل',noPeers:'لا توجد عُقد بعد. أضف عقدة موثوقة من إعدادات الاتصال.',daemon:'Core يستجيب',identity:'هوية العُقدة متاحة',tun:'واجهة TUN مفعّلة',peerHealth:'اتصال نقل واحد على الأقل متاح',up:'متصل',down:'غير متصل',inbound:'وارد',outbound:'صادر',saved:'تم حفظ الإعدادات وإعادة اتصال Core بنفس الهوية.',restarted:'تمت إعادة تشغيل Core.',busy:'جارٍ تطبيق التعديلات…',copied:'تم نسخ العنوان.',copyFailed:'حدّد العنوان وانسخه يدويًا.',advancedConfig:'توجد إعدادات اتصال متقدمة. أُوقف التعديل من المتصفح للحفاظ على الإعداد الأصلي.',restartConfirm:'إعادة تشغيل Core؟ سينقطع الاتصال لفترة قصيرة.',networkError:'لوحة الإدارة غير متاحة. حاول بعد قليل.'}
};
const byId = id => document.getElementById(id);
let language = localStorage.getItem('uqda-language') || (navigator.language.startsWith('ar') ? 'ar' : 'en');
if (!text[language]) language = 'en';
let csrf = '', snapshot = null, editing = false, applying = false, editRevision = '';
const t = key => text[language][key] || key;
function translate() {
  document.documentElement.lang = language;
  document.documentElement.dir = language === 'ar' ? 'rtl' : 'ltr';
  document.querySelectorAll('[data-i18n]').forEach(element => element.textContent = t(element.dataset.i18n));
  byId('language').textContent = language === 'ar' ? 'English' : 'العربية';
  if (snapshot) render(snapshot);
}
function notice(message, error = false) { byId('notice').textContent = message; byId('notice').className = error ? 'notice error' : 'notice'; byId('notice').hidden = !message; }
function authenticated(value) { byId('login').hidden = value; byId('dashboard').hidden = !value; byId('logout').hidden = !value; }
async function api(path, value) {
  let response;
  try { response = await fetch(path, value === undefined ? {cache:'no-store'} : {method:'POST',headers:{'Content-Type':'application/json','X-Uqda-CSRF':csrf},body:JSON.stringify(value)}); }
  catch (_) { throw new Error(t('networkError')); }
  const data = await response.json();
  if (!response.ok) { if (response.status === 401) authenticated(false); throw new Error(data.error || t('networkError')); }
  return data;
}
function modeFields() { const isPublic = byId('networkMode').value === 'public'; byId('privateFields').hidden = isPublic; byId('publicFields').hidden = !isPublic; byId('confirmPublic').required = isPublic; }
function render(data) {
  snapshot = data;
  const peers = data.peers || [], connected = peers.filter(peer => peer.up).length;
  byId('connection').textContent = t(!data.ready ? 'starting' : connected ? 'connected' : 'isolated');
  byId('connection').className = connected && data.ready ? 'badge up' : 'badge';
  byId('address').textContent = data.identity?.address || '—';
  byId('copyAddress').disabled = !data.identity?.address;
  byId('peerCount').textContent = connected;
  byId('interfaceName').textContent = data.tun?.enabled ? data.tun.name : '—';
  byId('version').textContent = data.identity?.build_version || '—';
  byId('modeValue').textContent = t(data.settings.private ? 'privateGroup' : 'publicNetwork');
  if (!editing) {
    editRevision = data.settings.revision;
    byId('peers').value = data.settings.peers.join('\n');
    byId('listeners').value = data.settings.listen.join('\n');
    byId('networkMode').value = data.settings.private ? 'private' : 'public';
    modeFields();
  }
  byId('settingsForm').querySelectorAll('input,textarea,select,button').forEach(element => element.disabled = !data.settings.editable || applying);
  if (!data.settings.editable) notice(t('advancedConfig'), true);
  byId('peerList').replaceChildren();
  if (!peers.length) { const empty = document.createElement('p'); empty.className = 'small empty'; empty.textContent = t('noPeers'); byId('peerList').append(empty); }
  for (const peer of peers) {
    const row = document.createElement('div'); row.className = 'peer';
    const head = document.createElement('div'); head.className = 'peer-head';
    const remote = document.createElement('code'); remote.textContent = peer.remote;
    const badge = document.createElement('span'); badge.className = peer.up ? 'badge up' : 'badge'; badge.textContent = t(peer.up ? 'up' : 'down');
    head.append(remote, badge);
    const meta = document.createElement('div'); meta.className = 'peer-meta'; meta.textContent = t(peer.inbound ? 'inbound' : 'outbound') + ' · RX ' + formatBytes(peer.bytes_recvd) + ' · TX ' + formatBytes(peer.bytes_sent);
    row.append(head, meta); byId('peerList').append(row);
  }
  byId('checks').replaceChildren();
  for (const [key, good] of [['daemon',data.ready],['identity',!!data.identity?.address],['tun',data.tun?.enabled],['peerHealth',connected > 0]]) {
    const li = document.createElement('li'), dot = document.createElement('span'); dot.className = good ? 'dot good' : 'dot';
    li.append(dot, document.createTextNode(t(key))); byId('checks').append(li);
  }
}
function formatBytes(value = 0) { const number = Number(value); if (number < 1024) return number + ' B'; if (number < 1048576) return (number / 1024).toFixed(1) + ' KB'; return (number / 1048576).toFixed(1) + ' MB'; }
async function refresh() { if (byId('dashboard').hidden || applying) return; try { render(await api('/api/status')); } catch (error) { notice(error.message, true); } }
byId('language').addEventListener('click', () => { language = language === 'en' ? 'ar' : 'en'; localStorage.setItem('uqda-language',language); translate(); });
byId('loginForm').addEventListener('submit', async event => { event.preventDefault(); const button = event.submitter; button.disabled = true; try { const data = await api('/api/login',{password:byId('loginPassword').value}); csrf = data.csrf; byId('loginPassword').value = ''; authenticated(true); notice(''); await refresh(); } catch (error) { notice(error.message, true); } finally { button.disabled = false; } });
byId('logout').addEventListener('click', async () => { try { await api('/api/logout',{}); csrf=''; authenticated(false); notice(''); } catch (error) { notice(error.message,true); } });
byId('settingsForm').addEventListener('input', () => { editing = true; });
byId('networkMode').addEventListener('change', modeFields);
byId('settingsForm').addEventListener('submit', async event => {
  event.preventDefault(); if (!snapshot || applying) return;
  const values = {revision:editRevision,peers:byId('peers').value.split('\n').map(value => value.trim()).filter(Boolean),listen:byId('listeners').value.split('\n').map(value => value.trim()).filter(Boolean),mode:byId('networkMode').value,group_password:byId('groupPassword').value,confirm_public:byId('confirmPublic').checked};
  applying = true; byId('save').disabled = true; byId('restart').disabled = true; notice(t('busy'));
  try { const data = await api('/api/settings',values); editing = false; byId('groupPassword').value = ''; byId('confirmPublic').checked = false; render(data); notice(t('saved')); }
  catch (error) { notice(error.message,true); }
  finally { applying = false; byId('restart').disabled = false; byId('save').disabled = !snapshot.settings.editable; }
});
byId('restart').addEventListener('click', async () => { if (applying || !confirm(t('restartConfirm'))) return; applying=true; byId('restart').disabled=true; notice(t('busy')); try { render(await api('/api/restart',{})); notice(t('restarted')); } catch (error) { notice(error.message,true); } finally { applying=false; byId('restart').disabled=false; } });
byId('refresh').addEventListener('click',refresh);
byId('copyAddress').addEventListener('click', async () => { try { await navigator.clipboard.writeText(byId('address').textContent); notice(t('copied')); } catch (_) { notice(t('copyFailed')); } });
translate();
(async () => { try { const session = await api('/api/session'); csrf=session.csrf; authenticated(session.authenticated); await refresh(); } catch (error) { notice(error.message,true); } })();
setInterval(refresh,10000);
