const usersContainer = document.getElementById('usersContainer');
const userDetail = document.getElementById('userDetail');
const refreshUsers = document.getElementById('refreshUsers');

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

function renderUsers(users) {
  usersContainer.innerHTML = '';
  if (!users.length) {
    usersContainer.innerHTML = '<p class="hint">Пользователей пока нет.</p>';
    return;
  }

  users.forEach((user) => {
    const card = document.createElement('div');
    card.className = 'link-card';
    card.innerHTML = `
      <div>
        <h3>${user.email}</h3>
        <p class="hint">Роль: ${user.role || 'user'}</p>
        <p class="hint">Создан: ${formatDate(user.created_at)}</p>
      </div>
      <div class="admin-actions">
        <button class="ghost" data-action="view">Смотреть</button>
        <button class="ghost" data-action="delete">Удалить</button>
      </div>
    `;

    card.querySelector('[data-action="view"]').addEventListener('click', () => loadUser(user.id));
    card.querySelector('[data-action="delete"]').addEventListener('click', () => deleteUser(user.id));

    usersContainer.appendChild(card);
  });
}

function renderUserDetail(detail) {
  userDetail.innerHTML = `
    <div class="sub-card">
      <div>
        <h3>${detail.user.email}</h3>
        <p class="hint">Роль: ${detail.user.role || 'user'}</p>
        <p class="hint">Подписка: ${detail.subscription.status || 'trial'}</p>
        <p class="hint">Пробный период до: ${formatDate(detail.subscription.trial_ends_at)}</p>
        <p class="hint">Оплачен до: ${formatDate(detail.subscription.current_period_end)}</p>
      </div>
    </div>
  `;

  if (!detail.links.length) {
    userDetail.innerHTML += '<p class="hint">Ссылок нет.</p>';
    return;
  }

  const list = document.createElement('ul');
  list.className = 'click-list';
  detail.links.forEach((link) => {
    const item = document.createElement('li');
    item.innerHTML = `<strong>${link.code}</strong> — ${link.target_url}`;
    list.appendChild(item);
  });
  userDetail.appendChild(list);
}

async function loadUsers() {
  try {
    const users = await fetchJSON('/api/admin/users');
    renderUsers(users);
  } catch (error) {
    usersContainer.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
}

async function loadUser(id) {
  try {
    const detail = await fetchJSON(`/api/admin/users/${id}`);
    renderUserDetail(detail);
  } catch (error) {
    userDetail.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
}

async function deleteUser(id) {
  if (!window.confirm('Удалить пользователя?')) {
    return;
  }
  try {
    await fetchJSON(`/api/admin/users/${id}`, { method: 'DELETE' });
    await loadUsers();
    userDetail.innerHTML = '<p class="hint">Пользователь удалён.</p>';
  } catch (error) {
    userDetail.innerHTML = `<p class="hint error">${error.message}</p>`;
  }
}

refreshUsers.addEventListener('click', loadUsers);

loadUsers();
