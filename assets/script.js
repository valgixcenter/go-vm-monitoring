var currentProcesses = [];
var sortCol = 'cpu';
var sortDesc = true;

document.addEventListener('DOMContentLoaded', () => {
    fetchStats();
    setInterval(fetchStats, 3000);
});

async function fetchStats() {
    try {
        const response = await fetch('/api/stats');
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        const data = await response.json();
        updateUI(data);
    } catch (error) {
        console.error('Error fetching stats:', error);
    }
}

function updateUI(data) {
    // CPU
    updateMetric('cpu-val', data.cpu_usage);
    updateBar('cpu-bar', data.cpu_usage);
    if (data.cpu_model) document.getElementById('cpu-model').textContent = data.cpu_model;
    if (data.cpu_cores && data.cpu_threads) {
        document.getElementById('cpu-cores').textContent = `${data.cpu_cores} cores / ${data.cpu_threads} threads`;
    }

    // Memory
    updateMetric('mem-val', data.memory_usage);
    updateBar('mem-bar', data.memory_usage);
    document.getElementById('mem-used').textContent = formatBytes(data.memory_used);
    document.getElementById('mem-total').textContent = formatBytes(data.memory_total);
    if (data.memory_type) document.getElementById('mem-type').textContent = data.memory_type;

    // Disk
    updateMetric('disk-val', data.disk_usage);
    updateBar('disk-bar', data.disk_usage);
    document.getElementById('disk-used').textContent = formatBytes(data.disk_used);
    document.getElementById('disk-total').textContent = formatBytes(data.disk_total);
    if (data.disk_fstype) document.getElementById('disk-fstype').textContent = data.disk_fstype;

    // Network
    document.getElementById('net-in').textContent = formatBytes(data.net_in_rate, true);
    document.getElementById('net-out').textContent = formatBytes(data.net_out_rate, true);

    // Processes
    if (data.processes) {
        currentProcesses = data.processes;
        renderTable();
    }
}

function updateMetric(id, value) {
    const el = document.getElementById(id);
    if (el) el.textContent = value;
}

function updateBar(id, value) {
    const el = document.getElementById(id);
    if (el) el.style.width = `${value}%`;

    // Change color based on load
    if (value > 90) {
        el.style.backgroundColor = 'var(--danger-color)';
        el.style.boxShadow = '0 0 10px rgba(248, 113, 113, 0.5)';
    } else if (value > 70) {
        el.style.backgroundColor = 'var(--warning-color)';
        el.style.boxShadow = '0 0 10px rgba(251, 191, 36, 0.5)';
    } else {
        el.style.backgroundColor = ''; // Reset to gradient
        el.style.boxShadow = '';
    }
}

function formatBytes(bytes, isRate = false) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i] + (isRate ? '/s' : '');
}

function sortTable(col) {
    if (sortCol === col) {
        sortDesc = !sortDesc;
    } else {
        sortCol = col;
        sortDesc = true;
    }
    renderTable();
}

function renderTable() {
    const tbody = document.getElementById('process-list');
    tbody.innerHTML = '';

    const sorted = [...currentProcesses].sort((a, b) => {
        let valA, valB;
        switch (sortCol) {
            case 'pid': valA = a.pid; valB = b.pid; break;
            case 'name': valA = a.name.toLowerCase(); valB = b.name.toLowerCase(); break;
            case 'cpu': valA = a.cpu_percent; valB = b.cpu_percent; break;
            case 'mem': valA = a.memory_percent; valB = b.memory_percent; break;
        }

        if (valA < valB) return sortDesc ? 1 : -1;
        if (valA > valB) return sortDesc ? -1 : 1;
        return 0;
    });

    sorted.forEach(p => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${p.pid}</td>
            <td>${p.name}</td>
            <td>${p.cpu_percent}%</td>
            <td>${p.memory_percent}% <span style="color: var(--text-secondary); font-size: 0.85em;">(${formatBytes(p.memory_rss)})</span></td>
        `;
        tbody.appendChild(tr);
    });
}
