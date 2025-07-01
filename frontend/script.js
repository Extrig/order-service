function fetchOrder() {
    const id = document.getElementById("orderIdInput").value.trim();
    if (!id) return alert("Введите Order UID");

    fetch(`/order/${id}`)
        .then(res => {
            if (!res.ok) throw new Error("Заказ не найден");
            return res.json();
        })
        .then(data => {
            renderInfo(data);
            renderJSON(data);
            showTab("info");
        })
        .catch(err => alert("Ошибка: " + err.message));
}

function renderInfo(order) {
    const infoEl = document.getElementById("infoTab");
    infoEl.innerHTML = `
    <h2>Общая информация</h2>
    <table>
      <tr><th>ID Заказа</th><td>${order.order_uid}</td></tr>
      <tr><th>Трек номер</th><td>${order.track_number}</td></tr>
      <tr><th>Сервис доставки</th><td>${order.delivery_service}</td></tr>
      <tr><th>Дата</th><td>${order.date_created}</td></tr>
    </table>

    <h2>Получатель</h2>
    <table>
      <tr><th>Имя</th><td>${order.delivery?.name || ''}</td></tr>
      <tr><th>Город</th><td>${order.delivery?.city || ''}</td></tr>
      <tr><th>Адрес</th><td>${order.delivery?.address || ''}</td></tr>
      <tr><th>Телефон</th><td>${order.delivery?.phone || ''}</td></tr>
      <tr><th>Почта</th><td>${order.delivery?.email || ''}</td></tr>
    </table>

    <h2>Оплата</h2>
    <table>
      <tr><th>Сумма</th><td>${order.payment?.amount || 0}</td></tr>
      <tr><th>Банк</th><td>${order.payment?.bank || ''}</td></tr>
      <tr><th>Метод</th><td>${order.payment?.provider || ''}</td></tr>
    </table>

    <h2>Товары</h2>
    <table>
      <tr><th>Название</th><th>Цена</th><th>Скидка</th><th>Итого</th></tr>
      ${order.items.map(item => `
        <tr>
          <td>${item.brand + ' ' + item.name}</td>
          <td>${item.price}</td>
          <td>${item.sale}%</td>
          <td>${item.total_price}</td>
        </tr>
      `).join('')}
    </table>
  `;
}

function renderJSON(data) {
    document.getElementById("jsonTab").textContent = JSON.stringify(data, null, 2);
}

function showTab(tab) {
    document.querySelectorAll(".tab-content").forEach(el => el.classList.remove("active"));
    document.querySelectorAll(".tab-button").forEach(btn => btn.classList.remove("active"));

    document.getElementById(tab + "Tab").classList.add("active");
    document.querySelector(`.tab-button[onclick*="${tab}"]`).classList.add("active");
}
