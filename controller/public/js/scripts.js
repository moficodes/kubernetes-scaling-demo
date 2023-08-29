const form = document.getElementById('form');
let data;
form.addEventListener('submit', handleSubmit);
const sendButton = document.getElementById('send');

async function handleSubmit(event) {
    event.preventDefault();
    const form = event.currentTarget;
    const url = new URL(form.action);
    const formData = new FormData(form);

    const fetchOptions = {
        method: form.method,
        body: formData,
    };

    const response = await fetch(url, fetchOptions);
    const json = await response.json();
    data = json.pixels;
    renderPixels(data);
    form.reset();
}

function renderPixels(data) {
    for (let i = 0; i < data.length; i++) {
        let row = "";
        for(j = 0; j < data[i].length; j++) {
            let r = data[i][j] >> 16 & 0xFF;
            let g = data[i][j] >> 8 & 0xFF;
            let b = data[i][j] & 0xFF;
            let color = `rgb(${r}, ${g}, ${b}) `;
            row += color;
        }
    }
    const pixelContainer = document.getElementById('pixels');
    pixelContainer.innerHTML = '';
    for(let i = 0; i < data.length; i++) {
        let row = document.createElement('div');
        row.classList.add('row');
        row.id = `row-${i}`
        for(j = 0; j < data[i].length; j++) {
            let column = document.createElement('div');
            column.id = `column-${i}-${j}`;
            column.classList.add('column');
            color = data[i][j];
            column.style.backgroundColor = "#" + color.toString(16).padStart(6, '0');;
            row.appendChild(column);
        }
        pixelContainer.appendChild(row);
    }   
}

async function sendPixels() {
    const url = "/direct"
    const body = {
        pixels: data
    }

    console.log(JSON.stringify(body));

    const fetchOptions = {
        method: "POST",
        body: JSON.stringify(body),
        Headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
        }
    }

    const response = await fetch(url, fetchOptions)
    const json = await response.json();
    console.log(json);
}
