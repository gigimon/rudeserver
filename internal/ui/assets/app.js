const listEl = document.getElementById('requestList');
const statsEl = document.getElementById('stats');
const detailTitle = document.getElementById('detailTitle');
const detailSubtitle = document.getElementById('detailSubtitle');
const detailStatus = document.getElementById('detailStatus');
const tabs = document.getElementById('tabs');

const filters = {
  search: document.getElementById('search'),
  method: document.getElementById('method'),
  status: document.getElementById('status'),
  jsonOnly: document.getElementById('jsonOnly'),
};

let currentId = null;
let currentList = [];
let sessionTotal = 0;

function statusBadge(status) {
  if (status >= 500) return 'bad';
  if (status >= 400) return 'warn';
  if (status >= 200) return 'good';
  return '';
}

function applyFilters(items) {
  const q = filters.search.value.toLowerCase();
  const method = filters.method.value;
  const statusGroup = filters.status.value;
  const jsonOnly = filters.jsonOnly.checked;

  return items.filter(item => {
    if (method && item.method !== method) return false;
    if (q && !(item.path + item.method).toLowerCase().includes(q)) return false;
    if (statusGroup) {
      const code = item.status || 0;
      if (statusGroup === '2xx' && (code < 200 || code >= 300)) return false;
      if (statusGroup === '4xx' && (code < 400 || code >= 500)) return false;
      if (statusGroup === '5xx' && (code < 500)) return false;
    }
    if (jsonOnly && !(item.content_type || '').includes('json')) return false;
    return true;
  });
}

function renderList() {
  const filtered = applyFilters(currentList);
  listEl.innerHTML = '';
  filtered.forEach(item => {
    const li = document.createElement('li');
    li.className = 'request-item' + (item.id === currentId ? ' active' : '');
    li.dataset.id = item.id;
    li.innerHTML = `
      <div class="request-meta">
        <span>${new Date(item.timestamp).toLocaleTimeString()}</span>
        <span class="badge ${statusBadge(item.status)}">${item.status}</span>
      </div>
      <div class="request-path">${item.method} ${item.path}</div>
      <div class="request-meta">
        <span>${item.client_ip || ''}</span>
        <span>${item.duration_ms || 0} ms</span>
      </div>
    `;
    li.addEventListener('click', () => loadDetail(item.id));
    listEl.appendChild(li);
  });

  const errors = currentList.filter(x => x.status >= 400).length;
  statsEl.innerHTML = `
    <div class="stat"><span>${currentList.length}</span><label>Visible</label></div>
    <div class="stat"><span>${errors}</span><label>Errors</label></div>
    <div class="stat"><span>${sessionTotal}</span><label>Session Total</label></div>
  `;
}

function setActiveTab(name) {
  document.querySelectorAll('.tab').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.tab === name);
  });
  document.querySelectorAll('.tab-content').forEach(el => {
    el.classList.toggle('hidden', el.id !== `tab-${name}`);
  });
}

tabs.addEventListener('click', e => {
  if (e.target.classList.contains('tab')) {
    setActiveTab(e.target.dataset.tab);
  }
});

async function loadList() {
  const res = await fetch('/ui/api/requests');
  if (!res.ok) return;
  const payload = await res.json();
  sessionTotal = payload.total || 0;
  currentList = payload.items || [];
  renderList();

  if (currentId && !currentList.find(x => x.id === currentId)) {
    currentId = null;
  }
}

async function loadDetail(id) {
  currentId = id;
  renderList();
  const res = await fetch(`/ui/api/requests/${id}`);
  if (!res.ok) return;
  const detail = await res.json();

  detailTitle.textContent = `${detail.method} ${detail.path}`;
  detailSubtitle.textContent = `${detail.protocol || ''} · ${detail.client_ip || ''}`;
  detailStatus.textContent = detail.status || '—';
  detailStatus.className = `pill badge ${statusBadge(detail.status)}`;

  document.getElementById('tab-overview').innerHTML = `
    <div class="kv-grid">
      <div>Status</div><div>${detail.status}</div>
      <div>Duration</div><div>${detail.duration_ms} ms</div>
      <div>User Agent</div><div>${detail.user_agent || ''}</div>
      <div>Query</div><div>${detail.query || ''}</div>
    </div>
  `;

  const renderHeaders = (headers = {}) => Object.entries(headers)
    .map(([k, v]) => `<div>${k}</div><div>${Array.isArray(v) ? v.join(', ') : v}</div>`)
    .join('');

  const pretty = (text) => {
    try {
      return JSON.stringify(JSON.parse(text), null, 2);
    } catch {
      return text;
    }
  };

  document.getElementById('tab-request').innerHTML = `
    <h3>Headers</h3>
    <div class="kv-grid">${renderHeaders(detail.req_headers)}</div>
    <h3>Body</h3>
    <div class="code-block">${pretty(detail.req_body || '')}</div>
  `;

  document.getElementById('tab-response').innerHTML = `
    <h3>Headers</h3>
    <div class="kv-grid">${renderHeaders(detail.res_headers)}</div>
    <h3>Body</h3>
    <div class="code-block">${pretty(detail.res_body || '')}</div>
  `;

  const rawBody = detail.req_utf8 ? detail.req_body : detail.req_b64;
  const rawRes = detail.res_utf8 ? detail.res_body : detail.res_b64;
  document.getElementById('tab-raw').innerHTML = `
    <h3>Request Raw</h3>
    <div class="code-block">${rawBody || ''}</div>
    <button class="copy-btn" data-copy="req">Copy</button>
    <h3>Response Raw</h3>
    <div class="code-block">${rawRes || ''}</div>
    <button class="copy-btn" data-copy="res">Copy</button>
  `;

  document.querySelectorAll('.copy-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const text = btn.dataset.copy === 'req' ? (rawBody || '') : (rawRes || '');
      navigator.clipboard.writeText(text);
    });
  });
}

Object.values(filters).forEach(input => {
  input.addEventListener('input', renderList);
});

setActiveTab('overview');
loadList();
setInterval(loadList, 4000);
