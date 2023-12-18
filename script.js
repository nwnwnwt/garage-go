document.addEventListener('DOMContentLoaded', function () {
    const addCarForm = document.getElementById('addCarForm');
    const carListContainer = document.getElementById('carList');

    addCarForm.addEventListener('submit', function (event) {
        event.preventDefault();

        const brand = document.getElementById('brand').value;
        const model = document.getElementById('model').value;

        fetch('/cars', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ brand, model }),
        })
        .then(response => response.json())
        .then(data => {
            alert(data.message);
            fetchCarList();
        })
        .catch(error => console.error('Error:', error));
    });

    function fetchCarList() {
        fetch('/cars')
        .then(response => response.json())
        .then(data => {
            const carList = data.cars;
            renderCarList(carList);
        })
        .catch(error => console.error('Error:', error));
    }

    function renderCarList(cars) {
        carListContainer.innerHTML = '';

        cars.forEach(car => {
            const carItem = document.createElement('div');
            carItem.classList.add('carItem');
            carItem.textContent = `${car.brand} - ${car.model} (${car.status})`;
            carListContainer.appendChild(carItem);
        });
    }

    // Fetch and render the initial car list
    fetchCarList();
});
