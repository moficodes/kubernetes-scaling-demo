const form = document.getElementById('form');
const sendButton = document.getElementById('send');
const colorPicker = document.getElementById('color');

let activeColor = "#FFFFFF";

colorPicker.addEventListener('change', (event) => {
    activeColor = event.target.value;
});

form.addEventListener('submit', handleSubmit);

let intervalID;

function setupBoard() {
    const pixelContainer = document.getElementById('pixels');
    console.log(pixelContainer);
    for (let i = 0; i < 64; i++) {
        let row = document.createElement('div');
        row.classList.add('row');
        row.id = `row-${i}`
        for (j = 0; j < 64; j++) {
            let column = document.createElement('div');
            column.id = `column-${i}-${j}`;
            column.classList.add('column');
            column.addEventListener('click', (event) => {
                event.target.style.backgroundColor = activeColor;
            });
            row.appendChild(column);
        }
        pixelContainer.appendChild(row);
    }
}

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
    renderPixels(json.pixels);
    form.reset();
}

function renderPixels(data) {
    for (let i = 0; i < data.length; i++) {
        for (let j = 0; j < data[i].length; j++) {
            let id = `column-${i}-${j}`;
            color = data[i][j];
            let column = document.getElementById(id);
            column.style.backgroundColor = "#" + color.toString(16).padStart(6, '0');
        }
    }
}

function hexToInteger(hex) {
    // Remove the leading hash sign, if any.
    if (hex.startsWith("#")) {
        hex = hex.substring(1);
    }

    // Convert the hexadecimal string to an integer.
    return parseInt(hex, 16);
}

function rgbToInt(rgb) {
    // Remove the leading "rgb(" and trailing ")".
    rgb = rgb.substring(4, rgb.length - 1);

    // Split the RGB values into an array.
    const rgbValues = rgb.split(",");

    // Convert each RGB value to an integer.
    const red = parseInt(rgbValues[0]);
    const green = parseInt(rgbValues[1]);
    const blue = parseInt(rgbValues[2]);

    // Calculate the integer value of the color.
    const colorInteger = (red << 16) + (green << 8) + blue;

    return colorInteger;
}

function getData() {
    let data = [];
    for (let i = 0; i < 64; i++) {
        let row = [];
        for (j = 0; j < 64; j++) {
            let column = document.getElementById(`column-${i}-${j}`);
            const computedStyle = window.getComputedStyle(column);
            const rgbColor = computedStyle.backgroundColor;
            color = rgbToInt(rgbColor);
            row.push(color);
        }
        data.push(row);
    }
    return data;
}

async function gameOfLife() {
    intervalID = setInterval(async () => {
        const url = "/gameoflife"

        const body = {
            pixels: getData(),
        }

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
        renderPixels(json.pixels);
        await sendPixels();
    }, 1000);
}

function stopGOL() {
    clearInterval(intervalID);
    intervalID = null;
}

async function sendPixels() {
    const url = "/direct"

    const body = {
        pixels: getData(),
    }

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

function clearPixels() {
    for (let i = 0; i < 64; i++) {
        for (j = 0; j < 64; j++) {
            let column = document.getElementById(`column-${i}-${j}`);
            column.style.backgroundColor = "#000000";
        }
    }
}

window.onload = function () {
    setupBoard();
}