const userEmail = document.getElementById('userEmail');
const logoutButton = document.getElementById('logoutButton');
const createLinkForm = document.getElementById('createLinkForm');
const linksContainer = document.getElementById('linksContainer');
const statsContainer = document.getElementById('statsContainer');
const refreshLinks = document.getElementById('refreshLinks');
const subscriptionStatus = document.getElementById('subscriptionStatus');
const renewButton = document.getElementById('renewButton');
const domainLabel = document.getElementById('domainLabel');

let appDomain = window.location.host;

function getToken() {
  return localStorage.getItem('token');
}

async function fetchJSON(url, options = {}) {
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${getToken()}`,
      ...(options.headers || {}),
    },
    ...options,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || 'Ошибка запроса');
  }
  return data;
}

function formatDate(value) {
  if (!value) return '—';
  const date = new Date(value);
  return date.toLocaleString('ru-RU');
}

function renderLinks(links) {
  linksContainer.innerHTML = '';
  if (!links.length) {
    linksContainer.innerHTML = '<p class="hint">У вас пока нет ссылок.</p>';
    return;
  }

  links.forEach((link) => {
    const card = document.createElement('div');
    card.className = 'link-card';

    const inviteUrl = `${appDomain}/i/${link.code}`;

    card.innerHTML = `
      <div>
        <h3>${inviteUrl}</h3>
        <p class="hint">Целевая: ${link.target_url}</p>
        <p class="hint">Создано: ${formatDate(link.created_at)}</p>
      </div>
      <button class="ghost">Статистика</button>
    `;

    card.querySelector('button').addEventListener('click', () => loadStats(link.id));

    linksContainer.appendChild(card);
  });
}

function renderStats(stats) {
  statsContainer.innerHTML = '';
  const header = document.createElement('div');
  header.className = 'stats-header';
  header.innerHTML = `
    <div>
      <h3>${appDomain}/i/${stats.code}</h3>
      <p class="hint">Целевая: ${stats.target}</p>
    </div>
    <div class="badge">Всего кликов: ${stats.total}</div>
  `;
  statsContainer.appendChild(header);

  if (!stats.clicks.length) {
    statsContainer.innerHTML += '<p class="hint">Пока нет переходов.</p>';
    return;
  }

  const list = document.createElement('ul');
  list.className = 'click-list';
  stats.clicks.forEach((click) => {
    const item = document.createElement('li');
    item.textContent = formatDate(click.clicked_at);
    list.appendChild(item);
  });
  statsContainer.appendChild(list);
}

function renderSubscription(subscription) {
  subscriptionStatus.innerHTML = `
    <div class="sub-card">
      <div>
        <h3>Статус: ${subscription.status}</h3>
        <p class="hint">Пробный период до: ${formatDate(subscription.trial_ends_at)}</p>
        <p class="hint">Оплачен до: ${formatDate(subscription.current_period_end)}</p>
      </div>
      <div class="badge">${subscription.monthly_price} ₽/месяц</div>
    </div>
  `;
}

async function loadConfig() {
  const config = await fetch('/api/config').then((r) => r.json());
  if (config.app_domain) {
    appDomain = `https://${config.app_domain}`;
    domainLabel.textContent = config.app_domain;
  }
}

async function loadUser() {
  try {
    const data = await fetchJSON('/api/me');
    userEmail.textContent = data.email;
    renderSubscription(data.subscription || {});
  } catch (error) {
    window.location.href = '/';
  }
}

async function loadLinks() {
  try {
    const links = await fetchJSON('/api/links');
    renderLinks(links);
  } catch (error) {
    linksContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
}

async function loadStats(id) {
  try {
    const stats = await fetchJSON(`/api/links/${id}/stats`);
    renderStats(stats);
  } catch (error) {
    statsContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
}

createLinkForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const formData = new FormData(createLinkForm);
  try {
    await fetchJSON('/api/links', {
      method: 'POST',
      body: JSON.stringify({ target_url: formData.get('target') }),
    });
    createLinkForm.reset();
    await loadLinks();
  } catch (error) {
    statsContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
});

renewButton.addEventListener('click', async () => {
  try {
    const data = await fetchJSON('/api/billing/checkout', { method: 'POST' });
    if (data.checkout_url) {
      window.open(data.checkout_url, '_blank');
    }
    await loadUser();
  } catch (error) {
    subscriptionStatus.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
});

logoutButton.addEventListener('click', () => {
  localStorage.removeItem('token');
  window.location.href = '/';
});

refreshLinks.addEventListener('click', () => {
  loadLinks();
});

loadConfig().then(() => {
  loadUser();
  loadLinks();
});
