const userEmail = document.getElementById('userEmail');
const logoutButton = document.getElementById('logoutButton');
const createLinkForm = document.getElementById('createLinkForm');
const linksContainer = document.getElementById('linksContainer');
const refreshLinks = document.getElementById('refreshLinks');
const subscriptionStatus = document.getElementById('subscriptionStatus');
const renewButton = document.getElementById('renewButton');
const domainLabel = document.getElementById('domainLabel');
const adminLink = document.getElementById('adminLink');

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
  if (!value || value.startsWith('0001-01-01')) return '—';
  const date = new Date(value);
  return date.toLocaleString('ru-RU');
}

function setAccessState(isBlocked) {
  createLinkForm.querySelector('button').disabled = isBlocked;
  refreshLinks.disabled = isBlocked;
  if (isBlocked) {
    linksContainer.innerHTML = '<p class="hint error">Оплатите подписку, чтобы видеть и создавать ссылки.</p>';
  }
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
      <a class="ghost-link" href="/stats?id=${link.id}">Статистика</a>
    `;

    linksContainer.appendChild(card);
  });
}

function renderSubscription(subscription) {
  const statusLabel = subscription.status === 'trial' ? 'Тестовый период' : subscription.status;
  const trialLabel = subscription.trial_days_left > 0
    ? `Осталось ${subscription.trial_days_left} дн.`
    : 'Тестовый период завершён';

  subscriptionStatus.innerHTML = `
    <div class="sub-card">
      <div>
        <h3>Статус: ${statusLabel}</h3>
        <p class="hint">${trialLabel}</p>
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
    if (data.role === 'admin') {
      adminLink.classList.remove('hidden');
    }
  } catch (error) {
    window.location.href = '/login';
  }
}

async function loadLinks() {
  try {
    const links = await fetchJSON('/api/links');
    setAccessState(false);
    renderLinks(links);
  } catch (error) {
    if (error.message.includes('Оплатите подписку')) {
      setAccessState(true);
      return;
    }
    linksContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
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
    linksContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
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
  window.location.href = '/login';
});

refreshLinks.addEventListener('click', () => {
  loadLinks();
});

loadConfig().then(() => {
  loadUser();
  loadLinks();
});
