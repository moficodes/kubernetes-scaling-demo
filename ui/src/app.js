// Import the functions you need from the SDKs you need
import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";
import { getFirestore, doc, getDoc } from "firebase/firestore"

const leds = document.getElementById("leds");

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
    apiKey: "AIzaSyBvF6pmf9rnqdxxjs88JOli41-oJJLSbl4",
    authDomain: "mofilabs-next-demo-02.firebaseapp.com",
    projectId: "mofilabs-next-demo-02",
    storageBucket: "mofilabs-next-demo-02.appspot.com",
    messagingSenderId: "690199829387",
    appId: "1:690199829387:web:bb56b5bfc40a40f42e6d8e",
    measurementId: "G-01E8M2RQYJ"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);

const db = getFirestore(app);

async function getMapping(db) {
    const docRef = doc(db, "mapping", "data")
    const docSnap = await getDoc(docRef);
    if (docSnap.exists()) {
        const { data } = docSnap.data();
        console.log(data);
        return data;
    }

    console.log("no such document");

    return null;
}

function intToHexColor(intValue) {
    if (intValue < 0 || intValue > 16777215) {
        throw new Error('Invalid integer value for a color.');
    }

    let hexString = intValue.toString(16);

    while (hexString.length < 6) {
        hexString = '0' + hexString;
    }

    return '#' + hexString;
}


async function getLedData(db) {
    const docRef = doc(db, "gke", "data")
    const docSnap = await getDoc(docRef);
    if (docSnap.exists()) {
        const { data } = docSnap.data();
        console.log(data);
        const byteString = data._byteString.binaryString;
        let res = [];
        for (let i = 0; i < byteString.length; i += 3) {
            let color = byteString.charCodeAt(i) << 16 | byteString.charCodeAt(i + 1) << 8 | byteString.charCodeAt(i + 2);
            const hex = intToHexColor(color);
            res.push(hex);
        }
        return res;
    }

    console.log("no such document");

    return null;
}

function createLedBoard() {
    for (let i = 0; i < 64; i++) {
        let row = document.createElement('row');
        row.classList.add('row');
        for (let j = 0; j < 64; j++) {
            let led = document.createElement('div');
            led.setAttribute('id', `${64 * i + j}`);
            led.classList.add('cell');
            row.appendChild(led);
        }

        leds.appendChild(row);
    }
}

function populateLedBoard(ledData, mapping) {
    for (let i = 0; i < ledData.length; i++) {
        const id = i.toString(10);
        const led = document.getElementById(id);
        led.style.backgroundColor = `${ledData[mapping[i]]}`;
    }
}

async function run() {
    const mapping = await getMapping(db);
    createLedBoard();
    let timer = setInterval(async () => {
        const ledData = await getLedData(db);
        populateLedBoard(ledData, mapping)
    }, 3000);
}

run();
