let chart;
let metadata;

window.onload = async function () {
    const response = await fetch('/metadata');
    metadata = await response.json();
}

async function generateLoad() {
    if (chart) {
        chart.destroy();
    }
    document.getElementById('load').disabled = true;
    const response = await fetch('/generate', {
        method: 'POST',
    });
    const json = await response.json();
    console.log(json)
    if (response.status === 200) {
        let req = document.getElementById('requests');
        req.innerHTML = json.requests;
        let avg = document.getElementById('average');
        avg.innerHTML = json.average;
        let labels = ["p50", "p90", "p95", "p99"];
        let codes = [json.p50, json.p90, json.p95, json.p99];

        let ctx = document.getElementById('chart');
        chart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Latency',
                    data: codes,
                    borderWidth: 1,
                }],
                options: {
                    scales: {
                        y: {
                            beginAtZero: true
                        },
                    }
                }
            }
        });

    }
    document.getElementById('load').disabled = false;

    if(metadata) {
        let metadataContainer = document.getElementById('links');
        metadata.forEach((item) => {
            if (item.type === 'youtube') {
                console.log(item);
                let frame = document.createElement('iframe');
                frame.title = item.title;
                frame.src = item.href;
                frame.classList.add('youtube');
                metadataContainer.appendChild(frame);
            } else {
                let div = document.createElement('div');
                div.classList.add('link');
                let link = document.createElement('a');
                link.href = item.href;
                link.target = '_blank';
                link.innerText = item.title;
                div.appendChild(link);
                metadataContainer.appendChild(div);
            }

        })
    }
}