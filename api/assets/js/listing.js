document.getElementById('addListingForm').addEventListener('submit', async e => {
  e.preventDefault();

  const $err = document.getElementById('alertError');
  const $ok = document.getElementById('alertSuccess');
  $err.style.display = 'none'; $ok.style.display = 'none';

  const title = document.getElementById('title').value.trim();
  const description = document.getElementById('description').value.trim();
  const address = document.getElementById('address').value.trim();
  const price = parseInt(document.getElementById('price').value, 10);
  const imageInput = document.getElementById('image');
  const file = imageInput.files[0];

  if (!file) {
    $err.textContent = 'Выберите картинку';
    $err.style.display = 'block';
    return;
  }

  if (file.size > 5 * 1024 * 1024) {
    $err.textContent = 'Размер картинки не должен превышать 5 МБ';
    $err.style.display = 'block';
    return;
  }

  const reader = new FileReader();
  reader.onload = async function() {
    const imageBase64 = reader.result.split(',')[1];

    try {
      const token = localStorage.getItem('AuthToken');
      const res = await fetch('/api/listings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'AuthToken': token
        },
        body: JSON.stringify({
          title,
          description,
          address,
          price,
          image_base64: imageBase64,
          image_name: file.name
        })
      });

      const result = await res.json();
      if (result.success) {
        $ok.textContent = 'Объявление создано!';
        $ok.style.display = 'block';
        setTimeout(() => window.location.href = '/', 1500);
      } else {
        throw new Error(result.message || 'Ошибка создания');
      }
    } catch (err) {
      $err.textContent = err.message || 'Ошибка';
      $err.style.display = 'block';
    }
  };

  reader.readAsDataURL(file);
});