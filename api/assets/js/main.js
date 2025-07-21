async function getAuthToken() {
  return localStorage.getItem('AuthToken') || '';
}

function updateHeaderButtons() {
  const token = localStorage.getItem('AuthToken');
  document.getElementById('loginBtn').style.display = token ? 'none' : '';
  document.getElementById('registerBtn').style.display = token ? 'none' : '';
  document.getElementById('logoutBtn').style.display = token ? '' : 'none';
  document.getElementById('addListingBtn').style.display = token ? '' : 'none';
}

let currentPage = 1;
let totalPages = 1;
let currentTargetUserId = '';

async function loadListings(page = 1) {
  const listingsDiv = document.getElementById('listings');
  const alertError = document.getElementById('alertError');
  const paginationDiv = document.getElementById('pagination');

  currentPage = page;
  listingsDiv.innerHTML = 'Загрузка...';
  alertError.style.display = 'none';

  const sortField = document.getElementById('sortField').value;
  const sortOrder = document.getElementById('sortOrder').value;
  const onlyLiked = document.getElementById('onlyLiked').checked;

  const minPrice = document.getElementById('minPrice').value;
  const maxPrice = document.getElementById('maxPrice').value;


  const params = new URLSearchParams({
    sort_field: sortField,
    sort_order: sortOrder,
    page: page,
    only_liked: onlyLiked,
    target_user_id: currentTargetUserId,
    min_price: minPrice || 0,
    max_price: maxPrice || 100000000
  });

  try {
    const token = await getAuthToken();
    const res = await fetch('/api/listings?' + params.toString(), {
      method: 'GET',
      headers: token ? { 'AuthToken': token } : {},
    });

    const result = await res.json();
    if (!result.success) throw new Error(result.message);

    listingsDiv.innerHTML = '';

    const listings = result.data.listings;
    if (listings.length === 0) {
      listingsDiv.innerHTML = 'Нет объявлений.';
      paginationDiv.innerHTML = '';
      return;
    }

    listings.forEach(listing => {
      const div = document.createElement('div');
      div.className = 'listing';

      const ownerButtons = listing.is_yours ? `
        <button class="edit-btn" data-id="${listing.id}">✏️ Редактировать</button>
        <button class="delete-btn" data-id="${listing.id}">🗑️ Удалить</button>
      ` : '';

      const liked = listing.is_liked;
      const likeButton = `
        <button class="like-btn" data-id="${listing.id}" data-liked="${liked}">
          ❤️ <span class="like-count">${listing.likes}</span>
        </button>
      `;

      div.innerHTML = `
        <h3>${listing.title}</h3>
        <img src="${listing.image_url}" alt="image" style="max-width:200px;max-height:200px;">
        <p>${listing.description}</p>
        <p>Адрес: ${listing.address}</p>
        <p>Цена: ${listing.price}</p>
        <p>Опубликовано: ${new Date(listing.created_at).toLocaleString()}</p>
        <p>Автор: <a href="#" class="author-link" data-id="${listing.author_id}">${listing.author_login || listing.author_id}</a></p>
        ${ownerButtons}
        <div>${likeButton}</div>
      `;

      listingsDiv.appendChild(div);
    });

    totalPages = result.data.total_pages || 1;
    renderPagination();

  } catch (err) {
    alertError.textContent = err.message;
    alertError.style.display = 'block';
  }
}

function renderPagination() {
  const paginationDiv = document.getElementById('pagination');
  paginationDiv.innerHTML = '';

  if (totalPages <= 1) return;

  if (currentPage > 1) {
    const prevBtn = document.createElement('button');
    prevBtn.className = 'page-btn';
    prevBtn.textContent = '←';
    prevBtn.onclick = () => loadListings(currentPage - 1);
    paginationDiv.appendChild(prevBtn);
  }

  const startPage = Math.max(1, currentPage - 2);
  const endPage = Math.min(totalPages, currentPage + 2);

  for (let i = startPage; i <= endPage; i++) {
    const pageBtn = document.createElement('button');
    pageBtn.className = `page-btn ${i === currentPage ? 'active' : ''}`;
    pageBtn.textContent = i;
    pageBtn.onclick = () => loadListings(i);
    paginationDiv.appendChild(pageBtn);
  }

  if (currentPage < totalPages) {
    const nextBtn = document.createElement('button');
    nextBtn.className = 'page-btn';
    nextBtn.textContent = '→';
    nextBtn.onclick = () => loadListings(currentPage + 1);
    paginationDiv.appendChild(nextBtn);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  updateHeaderButtons();

  document.getElementById('applyFilters').onclick = () => {
    currentTargetUserId = '';
    document.getElementById('filterInfo').style.display = 'none';
    loadListings(1);
  };

  document.getElementById('clearAuthorFilter').onclick = () => {
    currentTargetUserId = '';
    document.getElementById('filterInfo').style.display = 'none';
    loadListings(1);
  };

  document.addEventListener('click', (e) => {
    const authorLink = e.target.closest('.author-link');
    if (authorLink) {
      e.preventDefault();
      currentTargetUserId = authorLink.dataset.id;
      document.getElementById('filterAuthor').textContent = authorLink.textContent;
      document.getElementById('filterInfo').style.display = 'block';
      loadListings(1);
    }
  });

  document.getElementById('loginBtn').onclick = () => window.location.href = '/login';
  document.getElementById('registerBtn').onclick = () => window.location.href = '/register';
  document.getElementById('logoutBtn').onclick = async () => {
    const token = localStorage.getItem('AuthToken');
    if (token) {
      await fetch('/api/logout', {
        method: 'DELETE',
        headers: { 'AuthToken': token }
      });
      localStorage.removeItem('AuthToken');
      updateHeaderButtons();
      loadListings(1);
    }
  };
  document.getElementById('addListingBtn').onclick = () => window.location.href = '/listing';

  document.addEventListener('click', async (e) => {
    if (e.target.closest('.like-btn')) {
      const btn = e.target.closest('.like-btn');
      const listingId = btn.dataset.id;
      const liked = btn.dataset.liked === 'true';
      const token = await getAuthToken();
      const endpoint = liked ? '/api/removelike' : '/api/addlike';

      try {
        const res = await fetch(endpoint, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'AuthToken': token
          },
          body: JSON.stringify({ listing_id: listingId })
        });
        const result = await res.json();
        if (!result.success) throw new Error(result.message);

        const countSpan = btn.querySelector('.like-count');
        let count = parseInt(countSpan.textContent, 10);
        countSpan.textContent = liked ? count - 1 : count + 1;
        btn.dataset.liked = liked ? 'false' : 'true';
      } catch (err) {
        document.getElementById('alertError').textContent = err.message;
        document.getElementById('alertError').style.display = 'block';
      }
    }

    if (e.target.matches('.delete-btn')) {
      const listingId = e.target.dataset.id;
      if (confirm('Удалить объявление?')) {
        const token = await getAuthToken();
        try {
          const res = await fetch('/api/listings/' + listingId, {
            method: 'DELETE',
            headers: {
              'AuthToken': token
            }
          });
          const result = await res.json();
          if (!result.success) throw new Error(result.message);
          loadListings(currentPage);
        } catch (err) {
          document.getElementById('alertError').textContent = err.message;
          document.getElementById('alertError').style.display = 'block';
        }
      }
    }

    if (e.target.matches('.edit-btn')) {
      const listingId = e.target.dataset.id;
      window.location.href = `/edit?id=${listingId}`;
    }
  });

  loadListings(1);
});