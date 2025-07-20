function normalizeResponse(resp) {
  return {
    success:    resp.success,
    statusCode: resp.code,
    message:    resp.message,
    data:       resp.data
  };
}

document.getElementById('registerForm').addEventListener('submit', async e => {
  e.preventDefault();

  const $err = document.getElementById('alertError');
  const $ok = document.getElementById('alertSuccess');
  $err.style.display = 'none'; $ok.style.display = 'none';

  try {
    const username = document.getElementById('registerUsername').value.trim();
    const password = document.getElementById('registerPassword').value.trim();

    const [encUser, encPass] = await Promise.all([
      encryptData(username),
      encryptData(password)
    ]);

    const res = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: encUser, password: encPass })
    });

    const result = await res.json();

    if (result.success && result.data && result.data.AuthToken) {
      localStorage.setItem('AuthToken', result.data.AuthToken);
      window.location.href = '/';
    } else {
      throw new Error(result.message || 'Ошибка регистрации');
    }
  } catch (err) {
    $err.textContent = err.message || 'Unexpected error';
    $err.style.display = 'block';
    console.error(err);
  }
});