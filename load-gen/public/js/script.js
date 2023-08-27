let chart;
let metadata;
let env;

window.onload = async function () {
    const response = await fetch('/metadata');
    metadata = await response.json();
    const resp = await fetch('/environment');
    env = await resp.json();
    console.log(env);
    document.getElementById('environment').innerHTML = env.env;
}

async function generateLoad() {
    if (chart) {
        chart.destroy();
    }
    document.getElementById('load').disabled = true;
    document.getElementById('result').classList.add('invisible');
    document.getElementById('result').classList.remove('visible');
    document.getElementById('loader').classList.add('visible');
    document.getElementById('loader').classList.remove('invisible');
    const response = await fetch('/generate', {
        method: 'POST',
    });
    const json = await response.json();
    document.getElementById('loader').classList.add('invisible');
    document.getElementById('loader').classList.remove('visible');
    document.getElementById('result').classList.add('visible');
    document.getElementById('result').classList.remove('invisible');
    if (response.status === 200) {
        let req = document.getElementById('requests');
        req.innerHTML = json.requests;
        let avg = document.getElementById('average');
        avg.innerHTML = `${json.average} s`;
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
        metadataContainer.innerHTML = '';

        let p = document.createElement('p');
        p.innerText = 'Learn More: ';
        metadataContainer.appendChild(p);
        metadata.forEach((item) => {
            if (item.type === 'youtube') {
                console.log(item);
                // let div = document.createElement('div');
                // div.classList.add('youtube-container');
                let frame = document.createElement('iframe');
                frame.title = item.title;
                frame.src = item.href;
                frame.classList.add('youtube');

                // div.appendChild(frame);
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