'use strict';
const text = {
  en: {eyebrow:'YOUR UMBREL. CONNECTED.',heading:'A connection you control.',intro:'Manage your encrypted IPv6 connection from one place.',signOut:'Sign out',loginTitle:'Open your dashboard',loginHint:"Use the app password shown in Umbrel’s Uqda app details.",appPassword:'App password',signIn:'Sign in',nodeStatus:'Your node',nodeAddress:'Uqda IPv6 address',copy:'Copy',connectedPeers:'Connected peers',interface:'Network interface',coreVersion:'Core version',mode:'Access mode',accessHint:'A connected transport peer does not prove private-group connectivity. Services also need to listen on IPv6 and allow access through their firewall.',settings:'Connection settings',savedOnDevice:'Saved on this device',firstRun:'New nodes start isolated. Add a trusted peer and your shared group password to connect.',peerAddresses:'Peer addresses',peersHint:'One address per line. TLS, TCP, QUIC, WS and WSS are supported. No peer is added automatically.',privateGroup:'Private group',publicNetwork:'Public network',groupPassword:'Shared group password',groupHint:'Use the same strong secret on all members. Leave blank to keep the current secret. Stored secrets are never shown here.',publicConfirm:'I understand that public mode removes private-group protection. This can make services listening on my Uqda IPv6 address reachable by other public nodes.',advanced:'Incoming connections',listenerAddresses:'Listener addresses (optional)',listenerHint:'Leave empty for outbound-only connections. Use TLS, TCP or QUIC and a port from 1024 to 65535. Router/firewall changes are separate.',save:'Save and reconnect',restartHint:'Saving restarts Core. Your node identity stays the same.',peers:'Peers',refresh:'Refresh',health:'Local health',restart:'Restart Core',project:'Project & documentation ↗',connected:'Connected',isolated:'No connected peers',starting:'Starting',noPeers:'No peers yet. Add a trusted peer in connection settings.',daemon:'Core responding',identity:'Node identity available',tun:'TUN interface enabled',peerHealth:'At least one transport peer connected',up:'Connected',down:'Disconnected',inbound:'Incoming',outbound:'Outgoing',saved:'Settings saved. Core reconnected with the same identity.',restarted:'Core restarted.',busy:'Applying changes…',copied:'Address copied.',copyFailed:'Select the address and copy it manually.',advancedConfig:'Advanced peer settings detected. Browser editing is disabled to preserve the original configuration.',restartConfirm:'Restart Core? Connections will briefly disconnect.',networkError:'The dashboard is unavailable. Try again shortly.'},
  ar: {eyebrow:'جهاز UMBREL متصل بشبكتك',heading:'اتصال بإدارتك.',intro:'أدر اتصال IPv6 المشفّر من مكان واحد.',signOut:'تسجيل الخروج',loginTitle:'افتح لوحة الإدارة',loginHint:'استخدم كلمة مرور التطبيق الظاهرة في تفاصيل Uqda داخل Umbrel.',appPassword:'كلمة مرور التطبيق',signIn:'تسجيل الدخول',nodeStatus:'عُقدتك',nodeAddress:'عنوان Uqda IPv6',copy:'نسخ',connectedPeers:'العُقد المتصلة',interface:'واجهة الشبكة',coreVersion:'إصدار Core',mode:'وضع الوصول',accessHint:'اتصال النقل بعقدة أخرى لا يثبت اتصال المجموعة الخاصة. الخدمات تحتاج أيضًا للاستماع على IPv6 والسماح بالوصول عبر الجدار الناري.',settings:'إعدادات الاتصال',savedOnDevice:'محفوظة على هذا الجهاز',firstRun:'العُقد الجديدة تبدأ معزولة. أضف عقدة موثوقة وكلمة مرور مجموعتك المشتركة للاتصال.',peerAddresses:'عناوين العُقد',peersHint:'عنوان واحد في كل سطر. يدعم TLS وTCP وQUIC وWS وWSS. لا تُضاف أي عقدة تلقائيًا.',privateGroup:'مجموعة خاصة',publicNetwork:'شبكة عامة',groupPassword:'كلمة مرور المجموعة المشتركة',groupHint:'استخدم سرًا قويًا ومطابقًا لدى جميع الأعضاء. اترك الحقل فارغًا للحفاظ على السر الحالي. الأسرار المحفوظة لا تظهر هنا.',publicConfirm:'أفهم أن الوضع العام يزيل حماية المجموعة الخاصة، وقد يتيح للعُقد العامة الوصول إلى الخدمات التي تستمع على عنوان Uqda IPv6 الخاص بجهازي.',advanced:'الاتصالات الواردة',listenerAddresses:'عناوين الاستماع (اختياري)',listenerHint:'اتركه فارغًا للاتصالات الصادرة فقط. استخدم TLS أو TCP أو QUIC ومنفذًا بين 1024 و65535. إعداد الراوتر والجدار الناري منفصل.',save:'حفظ وإعادة الاتصال',restartHint:'الحفظ يعيد تشغيل Core ويحافظ على هوية العُقدة.',peers:'العُقد',refresh:'تحديث',health:'حالة الجهاز',restart:'إعادة تشغيل Core',project:'المشروع والتوثيق ↗',connected:'متصل',isolated:'لا توجد عُقد متصلة',starting:'جارٍ التشغيل',noPeers:'لا توجد عُقد بعد. أضف عقدة موثوقة من إعدادات الاتصال.',daemon:'Core يستجيب',identity:'هوية العُقدة متاحة',tun:'واجهة TUN مفعّلة',peerHealth:'اتصال نقل واحد على الأقل متاح',up:'متصل',down:'غير متصل',inbound:'وارد',outbound:'صادر',saved:'تم حفظ الإعدادات وإعادة اتصال Core بنفس الهوية.',restarted:'تمت إعادة تشغيل Core.',busy:'جارٍ تطبيق التعديلات…',copied:'تم نسخ العنوان.',copyFailed:'حدّد العنوان وانسخه يدويًا.',advancedConfig:'توجد إعدادات اتصال متقدمة. أُوقف التعديل من المتصفح للحفاظ على الإعداد الأصلي.',restartConfirm:'إعادة تشغيل Core؟ سينقطع الاتصال لفترة قصيرة.',networkError:'لوحة الإدارة غير متاحة. حاول بعد قليل.'}
};
const byId = id => document.getElementById(id);
Object.assign(text.en, {
  heading: 'Your services. Your connection.',
  intro: 'Reach your own Umbrel services over an encrypted IPv6 network. Set up access deliberately, one service at a time.',
  servicesTitle:'Your services over Uqda',noAutoPublish:'No automatic publishing',
  servicesPurpose:'Use your own files, web apps or SSH from another Uqda-connected device. Saving a service creates an address and instructions; it does not install the app or open its ports.',
  deviceStep:'1 · Connect your other device',deviceHelp:'Install Uqda on that device. Use compatible network modes; for a private group, use the same strong group password and a reachable peer. Opening this dashboard alone does not connect a phone or tablet to Uqda.',
  serviceStep:'2 · Prepare one service',serviceHelp:"Use a service already installed on Umbrel. Enable IPv6 listening and allow only the intended clients through its firewall. Keep the service’s own login enabled. IPv4-only Docker apps need separate configuration.",
  verifyStep:'3 · Verify from the other device',verifyHelp:'Check the network, then use the service command below. A local TCP check does not prove remote access, app health or authentication.',
  servicePublic:'You are in public mode. Other public nodes may reach IPv6 services if their firewall permits it. Prefer a private group for your own devices; never remove app authentication.',
  serviceName:'Service name',serviceKind:'Service type',servicePort:'Existing TCP port',addService:'Save service',
  httpWarning:'Prefer HTTPS. Plain HTTP does not add application-level TLS. Do not share the dashboard password or group secret in a service address.',
  serviceLimits:'This is a service address book, not a VPN internet exit, reverse proxy or app installer. No automatic discovery, routing or firewall changes. Copy web addresses into a separate browser profile; same-host HTTP cookies can be shared across ports.',
  noServices:'No services saved yet. Choose an installed service and its actual IPv6 TCP port. Saving it will not publish it.',
  checkLocal:'Check local TCP',removeService:'Remove entry',removeConfirm:'Remove this address-book entry? The installed service and its firewall settings will not change.',
  serviceSaved:'Service instructions saved. No ports were opened.',serviceRemoved:'Address-book entry removed. The installed service was not removed.',
  localUnknown:'Local TCP: not checked',localPass:'Last local TCP check: accepted a connection',localFail:'Last local TCP check: unavailable',
  remotePending:'Remote access: not verified — test from your other Uqda device.',networkCommand:'On your other device · network check',serviceCommand:'Then check or use the service',localExplanation:'This check connects only to this node’s own IPv6 TCP port; it does not check application health, TLS, authentication or remote access.',
  noServiceAddress:'Start Core to generate this service’s Uqda address.',sshUser:'SSH: replace USER with your existing account name. This does not create an account.',
  purposeTitle: 'Why use Uqda on Umbrel?',
  purposeText: 'Give your Umbrel host a persistent encrypted IPv6 address and manage trusted connections without editing configuration files. Use a private group to connect your own devices.',
  setupTitle: 'Before using a service',
  setupPeer: 'Connect a trusted peer. Private-group members must use the same strong group password.',
  setupTest: 'From another group member, run sudo uqda test followed by the address shown below. A green peer badge alone does not verify end-to-end connectivity.',
  setupService: 'Then test the intended service separately. It must listen on IPv6 and its firewall must allow access. Docker apps are not exposed automatically.',
  recoveryTitle: 'Keep your identity',
  recoveryText: 'Update and restart through Umbrel. Back up app data first: deleting it removes your node identity and group settings. Never restore one identity onto two active nodes.'
});
Object.assign(text.ar, {
  heading:'خدماتك. اتصال بإدارتك.',intro:'استخدم خدمات Umbrel الخاصة بك عبر شبكة IPv6 مشفّرة. جهّز الوصول بأمان، خدمة واحدة في كل مرة.',
  servicesTitle:'خدماتك عبر عقدة',noAutoPublish:'لا يوجد نشر تلقائي',
  servicesPurpose:'استخدم ملفاتك أو تطبيقات الويب أو SSH من جهاز آخر متصل بعقدة. حفظ الخدمة يولّد عنوانًا وإرشادات؛ لا يثبّت التطبيق ولا يفتح منافذه.',
  deviceStep:'١ · اربط جهازك الآخر',deviceHelp:'ثبّت عقدة عليه واستخدم وضع شبكة متوافقًا. للمجموعة الخاصة، استخدم كلمة المرور القوية نفسها وعقدة يمكن الوصول إليها. فتح هذه اللوحة وحده لا يربط الهاتف أو الآيباد بشبكة عقدة.',
  serviceStep:'٢ · جهّز خدمة واحدة',serviceHelp:'اختر خدمة مثبتة بالفعل على Umbrel. فعّل الاستماع على IPv6 واسمح فقط بالأجهزة المطلوبة عبر جدارها الناري. أبقِ تسجيل الدخول الخاص بالخدمة مفعّلًا. تطبيقات Docker التي تدعم IPv4 فقط تحتاج إعدادًا منفصلًا.',
  verifyStep:'٣ · تحقّق من الجهاز الآخر',verifyHelp:'افحص الشبكة ثم استخدم أمر الخدمة أدناه. الفحص المحلي لمنفذ TCP لا يثبت الوصول عن بُعد أو سلامة التطبيق أو تسجيل الدخول.',
  servicePublic:'أنت في الوضع العام. قد تصل العقد العامة الأخرى لخدمات IPv6 إذا سمح جدارها الناري. فضّل مجموعة خاصة لأجهزتك، ولا تلغِ تسجيل دخول التطبيقات.',
  serviceName:'اسم الخدمة',serviceKind:'نوع الخدمة',servicePort:'منفذ TCP الموجود',addService:'حفظ الخدمة',
  httpWarning:'فضّل HTTPS. استخدام HTTP لا يضيف TLS على مستوى التطبيق. لا تضع كلمة مرور اللوحة أو سر المجموعة في عنوان الخدمة.',
  serviceLimits:'هذا دليل عناوين للخدمات، وليس بوابة إنترنت VPN أو وكيلًا عكسيًا أو مثبّت تطبيقات. لا يكتشف التطبيقات ولا يغيّر التوجيه أو الجدار الناري تلقائيًا. انسخ عناوين الويب إلى ملف متصفح منفصل؛ قد تُشارك كوكيز HTTP بين منافذ المضيف نفسه.',
  noServices:'لا توجد خدمات محفوظة بعد. اختر خدمة مثبتة ومنفذ TCP الفعلي الذي يستمع على IPv6. حفظها لا ينشرها.',
  checkLocal:'فحص TCP المحلي',removeService:'حذف من الدليل',removeConfirm:'حذف هذا الإدخال من دليل العناوين؟ لن تتغير الخدمة المثبتة أو إعدادات جدارها الناري.',
  serviceSaved:'حُفظت إرشادات الخدمة. لم تُفتح أي منافذ.',serviceRemoved:'حُذف الإدخال من الدليل. لم تُحذف الخدمة المثبتة.',
  localUnknown:'TCP المحلي: لم يُفحص',localPass:'آخر فحص TCP محلي: قَبِل اتصالًا',localFail:'آخر فحص TCP محلي: غير متاح',
  remotePending:'الوصول عن بُعد: لم يُتحقق منه — اختبره من جهازك الآخر المتصل بعقدة.',networkCommand:'على جهازك الآخر · فحص الشبكة',serviceCommand:'ثم افحص الخدمة أو استخدمها',localExplanation:'يتصل الفحص فقط بمنفذ TCP على عنوان IPv6 لهذه العقدة؛ لا يفحص التطبيق أو TLS أو تسجيل الدخول أو الوصول عن بُعد.',
  noServiceAddress:'شغّل Core لتوليد عنوان عقدة لهذه الخدمة.',sshUser:'SSH: استبدل USER باسم حسابك الموجود. هذا الأمر لا ينشئ حسابًا.',
  purposeTitle: 'لماذا تستخدم Uqda على Umbrel؟',
  purposeText: 'امنح جهاز Umbrel عنوان IPv6 مشفّرًا ثابت الهوية، وأدر الاتصالات الموثوقة دون تعديل ملفات الإعدادات. استخدم مجموعة خاصة لربط أجهزتك ببعضها.',
  setupTitle: 'قبل استخدام أي خدمة',
  setupPeer: 'اتصل بعُقدة موثوقة. يجب أن تستخدم جميع عُقد المجموعة الخاصة كلمة مرور قوية ومطابقة.',
  setupTest: 'من جهاز آخر في مجموعتك، شغّل sudo uqda test ثم العنوان الظاهر أدناه. ظهور اتصال أخضر وحده لا يثبت الاتصال الكامل بين الجهازين.',
  setupService: 'اختبر الخدمة المطلوبة بشكل منفصل. يجب أن تستمع على IPv6 وأن يسمح جدارها الناري بالوصول. لا تُكشف تطبيقات Docker تلقائيًا.',
  recoveryTitle: 'حافظ على هويتك',
  recoveryText: 'حدّث التطبيق وأعد تشغيله من Umbrel. انسخ بياناته احتياطيًا أولًا: حذفها يزيل هوية العُقدة وإعدادات المجموعة. لا تستعد هوية واحدة على عُقدتين تعملان معًا.'
});
let language = localStorage.getItem('uqda-language') || (navigator.language.startsWith('ar') ? 'ar' : 'en');
if (!text[language]) language = 'en';
function readSessionProof() { try { return sessionStorage.getItem('uqda-session-proof') || ''; } catch (_) { return ''; } }
function storeSessionProof(value) { try { if (value) sessionStorage.setItem('uqda-session-proof', value); else sessionStorage.removeItem('uqda-session-proof'); } catch (_) {} }
let csrf = readSessionProof(), snapshot = null, editing = false, applying = false, editRevision = '', generation = 0;
let serviceBusy = false, serviceContext = '', serviceChecks = new Map();
const t = key => text[language][key] || key;
function translate() {
  document.documentElement.lang = language;
  document.documentElement.dir = language === 'ar' ? 'rtl' : 'ltr';
  document.querySelectorAll('[data-i18n]').forEach(element => element.textContent = t(element.dataset.i18n));
  byId('language').textContent = language === 'ar' ? 'English' : 'العربية';
  if (snapshot) render(snapshot);
}
function notice(message, error = false) { byId('notice').textContent = message; byId('notice').className = error ? 'notice error' : 'notice'; byId('notice').hidden = !message; }
function authenticated(value) {
  generation++;
  byId('login').hidden = value; byId('dashboard').hidden = !value; byId('logout').hidden = !value;
  if (!value) {
    csrf = ''; storeSessionProof(''); snapshot = null; editing = false; editRevision = '';
    byId('settingsForm').reset(); byId('loginForm').reset();
    byId('serviceForm').reset(); byId('serviceList').replaceChildren();
    serviceChecks.clear(); serviceContext = '';
    byId('address').textContent = '—'; byId('peerList').replaceChildren(); byId('checks').replaceChildren();
    modeFields();
  }
}
function setApplying(value) {
  applying = value;
  if (value) { generation++; serviceChecks.clear(); if (snapshot) renderServices(snapshot); }
  byId('settingsForm').querySelectorAll('input,textarea,select,button').forEach(element => element.disabled = value || !snapshot?.settings.editable);
  byId('restart').disabled = value;
  byId('logout').disabled = value;
  serviceControls();
}
async function api(path, value) {
  let response;
  try { response = await fetch(path, value === undefined ? {cache:'no-store',headers:{'X-Uqda-CSRF':csrf}} : {method:'POST',headers:{'Content-Type':'application/json','X-Uqda-CSRF':csrf},body:JSON.stringify(value)}); }
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
  renderServices(data);
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
function serviceControls() {
  byId('serviceForm').querySelectorAll('input,select,button').forEach(element => element.disabled = serviceBusy || applying || !snapshot);
  byId('serviceList').querySelectorAll('button').forEach(element => element.disabled = serviceBusy || applying || !snapshot || (element.dataset.probe === 'true' && !snapshot.tun?.enabled));
  byId('restart').disabled = applying || serviceBusy;
  byId('save').disabled = applying || serviceBusy || !snapshot?.settings.editable;
  byId('logout').disabled = applying || serviceBusy;
}
function serviceLine(parent, label, value) {
  const paragraph = document.createElement('p'); paragraph.className = 'field-hint'; paragraph.textContent = label;
  const code = document.createElement('code'); code.textContent = value;
  parent.append(paragraph, code);
}
function renderServices(data) {
  const context = [data.identity?.address, data.services?.revision, data.settings.revision, data.ready, data.tun?.enabled].join('|');
  if (context !== serviceContext) { serviceChecks.clear(); serviceContext = context; }
  byId('serviceModeWarning').hidden = data.settings.private;
  byId('serviceList').replaceChildren();
  const items = data.services?.items || [];
  if (!items.length) { const empty = document.createElement('p'); empty.className = 'small empty'; empty.textContent = t('noServices'); byId('serviceList').append(empty); }
  for (const item of items) {
    const row = document.createElement('article'); row.className = 'service-entry';
    const title = document.createElement('h3'); title.textContent = item.name + ' · ' + item.kind.toUpperCase() + ' · ' + item.port; row.append(title);
    if (item.endpoint) {
      const endpoint = document.createElement('code'); endpoint.textContent = item.endpoint; row.append(endpoint);
      serviceLine(row, t('networkCommand'), item.network_command);
      serviceLine(row, t('serviceCommand'), item.service_command);
    } else { const hint = document.createElement('p'); hint.textContent = t('noServiceAddress'); row.append(hint); }
    if (item.kind === 'ssh') { const hint = document.createElement('p'); hint.className = 'field-hint'; hint.textContent = t('sshUser'); row.append(hint); }
    const result = serviceChecks.get(item.id);
    const badge = document.createElement('p'); badge.className = 'small'; badge.textContent = t(result === undefined ? 'localUnknown' : result ? 'localPass' : 'localFail');
    const pending = document.createElement('p'); pending.className = 'field-hint'; pending.textContent = t('remotePending');
    const explanation = document.createElement('p'); explanation.className = 'field-hint'; explanation.textContent = t('localExplanation');
    const actions = document.createElement('div'); actions.className = 'service-actions';
    const probe = document.createElement('button'); probe.type = 'button'; probe.dataset.probe = 'true'; probe.textContent = t('checkLocal'); probe.disabled = !data.tun?.enabled;
    probe.addEventListener('click', () => serviceRequest('/api/services/probe', {id:item.id}, true));
    const remove = document.createElement('button'); remove.type = 'button'; remove.className = 'quiet'; remove.textContent = t('removeService');
    remove.addEventListener('click', () => { if (confirm(t('removeConfirm'))) serviceRequest('/api/services/remove', {revision:data.services.revision,id:item.id}); });
    actions.append(probe, remove); row.append(badge, pending, explanation, actions); byId('serviceList').append(row);
  }
  serviceControls();
}
async function serviceRequest(path, values, probe = false) {
  if (serviceBusy || applying || !snapshot) return;
  serviceBusy = true; generation++; serviceControls();
  const current = generation, context = serviceContext;
  try {
    const data = await api(path, values);
    if (current !== generation || byId('dashboard').hidden) return;
    if (probe) {
      if (context === serviceContext) { serviceChecks.set(data.id, data.tcp_reachable); renderServices(snapshot); }
    } else {
      render(data);
      if (path.endsWith('/add')) byId('serviceForm').reset();
      notice(t(path.endsWith('/add') ? 'serviceSaved' : 'serviceRemoved'));
    }
  } catch (error) { if (current === generation) notice(error.message, true); }
  finally { serviceBusy = false; serviceControls(); }
}
function formatBytes(value = 0) { const number = Number(value); if (number < 1024) return number + ' B'; if (number < 1048576) return (number / 1024).toFixed(1) + ' KB'; return (number / 1048576).toFixed(1) + ' MB'; }
async function refresh() {
  if (byId('dashboard').hidden || applying || serviceBusy) return;
  const current = generation;
  try {
    const data = await api('/api/status');
    if (current === generation && !byId('dashboard').hidden && !applying && !serviceBusy) render(data);
  } catch (error) { if (current === generation) notice(error.message, true); }
}
byId('language').addEventListener('click', () => { language = language === 'en' ? 'ar' : 'en'; localStorage.setItem('uqda-language',language); translate(); });
byId('loginForm').addEventListener('submit', async event => { event.preventDefault(); const button = event.submitter; button.disabled = true; try { const data = await api('/api/login',{password:byId('loginPassword').value}); csrf = data.csrf; storeSessionProof(csrf); byId('loginPassword').value = ''; authenticated(true); notice(''); await refresh(); } catch (error) { notice(error.message, true); } finally { button.disabled = false; } });
byId('logout').addEventListener('click', async () => { try { await api('/api/logout',{}); csrf=''; authenticated(false); notice(''); } catch (error) { notice(error.message,true); } });
byId('settingsForm').addEventListener('input', () => { editing = true; });
byId('networkMode').addEventListener('change', modeFields);
byId('settingsForm').addEventListener('submit', async event => {
  event.preventDefault(); if (!snapshot || applying || serviceBusy) return;
  const values = {revision:editRevision,peers:byId('peers').value.split('\n').map(value => value.trim()).filter(Boolean),listen:byId('listeners').value.split('\n').map(value => value.trim()).filter(Boolean),mode:byId('networkMode').value,group_password:byId('groupPassword').value,confirm_public:byId('confirmPublic').checked};
  setApplying(true); notice(t('busy'));
  try { const data = await api('/api/settings',values); editing = false; byId('groupPassword').value = ''; byId('confirmPublic').checked = false; render(data); notice(t('saved')); }
  catch (error) { notice(error.message,true); }
  finally { setApplying(false); }
});
byId('restart').addEventListener('click', async () => { if (applying || serviceBusy || !confirm(t('restartConfirm'))) return; setApplying(true); notice(t('busy')); try { render(await api('/api/restart',{})); notice(t('restarted')); } catch (error) { notice(error.message,true); } finally { setApplying(false); } });
byId('serviceForm').addEventListener('submit', event => { event.preventDefault(); if (!snapshot) return; serviceRequest('/api/services/add', {revision:snapshot.services.revision,name:byId('serviceName').value,kind:byId('serviceKind').value,port:Number(byId('servicePort').value)}); });
byId('refresh').addEventListener('click',refresh);
byId('copyAddress').addEventListener('click', async () => { try { await navigator.clipboard.writeText(byId('address').textContent); notice(t('copied')); } catch (_) { notice(t('copyFailed')); } });
translate();
(async () => { try { const session = await api('/api/session'); csrf=session.csrf; authenticated(session.authenticated); await refresh(); } catch (error) { notice(error.message,true); } })();
setInterval(refresh,10000);
