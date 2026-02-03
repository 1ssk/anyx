const loginForm = document.getElementById('loginForm');
const authMessage = document.getElementById('authMessage');

function showMessage(text, isError = true) {
  authMessage.textContent = text;
  authMessage.classList.toggle('error', isError);
}

async function sendAuth(url, payload) {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || 'Ошибка авторизации');
  }
  localStorage.setItem('token', data.token);
}

loginForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const formData = new FormData(loginForm);
  try {
    await sendAuth('/api/login', {
      email: formData.get('email'),
      password: formData.get('password'),
    });
    window.location.href = '/dashboard';
  } catch (error) {
    showMessage(error.message);
  }
});
