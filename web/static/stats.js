const statsContainer = document.getElementById('statsContainer');

function getToken() {
  return localStorage.getItem('token');
}

function getLinkID() {
  const params = new URLSearchParams(window.location.search);
  return params.get('id');
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

function renderStats(stats) {
  statsContainer.innerHTML = '';
  const header = document.createElement('div');
  header.className = 'stats-header';
  header.innerHTML = `
    <div>
      <h3>${stats.code}</h3>
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

async function loadStats() {
  const id = getLinkID();
  if (!id) {
    statsContainer.innerHTML = '<p class="hint error">Не указан идентификатор ссылки.</p>';
    return;
  }

  try {
    const stats = await fetchJSON(`/api/links/${id}/stats`);
    renderStats(stats);
  } catch (error) {
    statsContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
}

loadStats();
