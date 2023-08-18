let chart;

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
        let statuses = Object.keys(json.statusCodes);
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
}