const API_KEY = 'access-control-secret';

async function api(path) {
  const res = await fetch(path, {
    headers: { 'X-API-Key': API_KEY }
  });
  if (!res.ok) {
    throw new Error('请求失败: ' + res.status);
  }
  return res.json();
}

function el(tag, cls, text) {
  const node = document.createElement(tag);
  if (cls) node.className = cls;
  if (text !== undefined) node.textContent = text;
  return node;
}

function renderCards(ov) {
  const items = [
    ['区域', ov.zone_count],
    ['门禁点', ov.access_point_count],
    ['读卡器', ov.reader_count],
    ['人员', ov.person_count],
    ['凭证', ov.credential_count],
    ['通行记录', ov.access_log_count],
    ['授权通过', ov.granted_count],
    ['拒绝', ov.denied_count],
    ['待处理告警', ov.open_alert_count]
  ];
  const wrap = document.getElementById('stats-cards');
  wrap.innerHTML = '';
  items.forEach(function (it) {
    const card = el('div', 'card');
    card.appendChild(el('div', 'num', String(it[1])));
    card.appendChild(el('div', 'label', it[0]));
    wrap.appendChild(card);
  });
}

function renderCredentialStats(stats) {
  const wrap = document.getElementById('credential-stats');
  wrap.innerHTML = '';
  (stats || []).forEach(function (s) {
    wrap.appendChild(el('span', 'chip', s.status + '：' + s.count));
  });
}

function renderZoneTraffic(items) {
  const tbody = document.querySelector('#zone-traffic tbody');
  tbody.innerHTML = '';
  (items || []).forEach(function (z) {
    const tr = document.createElement('tr');
    const cells = [z.zone_name, z.granted, z.denied, z.total];
    cells.forEach(function (c) { tr.appendChild(el('td', null, c)); });
    tbody.appendChild(tr);
  });
}

function renderAccessLogs(items) {
  const tbody = document.querySelector('#access-logs tbody');
  tbody.innerHTML = '';
  (items || []).forEach(function (l) {
    const tr = document.createElement('tr');
    tr.appendChild(el('td', null, l.id));
    tr.appendChild(el('td', null, l.access_point_id));
    tr.appendChild(el('td', null, l.direction));
    const resultTd = document.createElement('td');
    const tag = el('span', 'tag ' + (l.result === 'granted' ? 'ok' : 'bad'), l.result);
    resultTd.appendChild(tag);
    tr.appendChild(resultTd);
    tr.appendChild(el('td', null, l.reason || '-'));
    tr.appendChild(el('td', null, l.access_at));
    tbody.appendChild(tr);
  });
}

async function load() {
  try {
    const ov = (await api('/api/stats/overview')).data;
    renderCards(ov);
    renderCredentialStats(ov.credential_stats);

    const traffic = (await api('/api/stats/zone-traffic')).data;
    renderZoneTraffic(traffic);

    const logs = (await api('/api/access-logs?size=50')).data;
    renderAccessLogs(logs.items);
  } catch (err) {
    document.getElementById('stats-cards').textContent = '加载失败: ' + err.message;
  }
}

load();
