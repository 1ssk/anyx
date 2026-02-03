const registerForm = document.getElementById('registerForm');
const authMessage = document.getElementById('authMessage');
const trialDays = document.getElementById('trialDays');
const monthlyPrice = document.getElementById('monthlyPrice');

function showMessage(text, isError = true) {
  authMessage.textContent = text;
  authMessage.classList.toggle('error', isError);
}

async function fetchConfig() {
  const response = await fetch('/api/config');
  const data = await response.json().catch(() => ({}));
  if (data.trial_days) {
    trialDays.textContent = data.trial_days;
  }
  if (data.monthly_price) {
    monthlyPrice.textContent = data.monthly_price;
  }
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

registerForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const formData = new FormData(registerForm);
  try {
    await sendAuth('/api/register', {
      email: formData.get('email'),
      password: formData.get('password'),
    });
    window.location.href = '/dashboard';
  } catch (error) {
    showMessage(error.message);
  }
});

fetchConfig();
