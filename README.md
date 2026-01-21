# Go VM Monitor

A lightweight, modern system monitoring dashboard for Linux VPS/VMs, written in Go. 

It provides real-time metrics for CPU, Memory, Disk, and Network usage, along with a detailed process list and hardware information, all presented in a beautiful "Glassmorphism" UI.

![Go VM Monitor Hero](https://providercontent.valgix.com/img/modules/govmmonitoring/hero.png)

## Features

- **Real-time monitoring**: Polling every 3 seconds for up-to-date stats.
- **System metrics**:
  - **CPU**: Usage %, Model name, Core & Thread counts.
  - **Memory**: Usage %, Total/Used, and *DDR Type* (Best effort detection).
  - **Disk**: Usage % for Root partition, Total/Used, and Filesystem Type.
  - **Network**: Real-time Incoming/Outgoing traffic rates.
- **Process manager**:
  - Top 20 processes sorted by CPU usage.
  - Detailed stats: PID, Name, CPU %, Memory % and **Absolute Memory Usage** (e.g., 250MB).
- **JSON API**: Full system statistics available at `/api.json`.
- **Modern UI**: Dark-themed, responsive design with smooth animations.

## Installation

### Prerequisites
- **Go** (1.20+ recommended)
- **Linux** (Tested on Ubuntu/Debian)
- *Optional*: `dmidecode` (installed by default on most distros) for reading Memory Type.

### Metrics Library
This project uses `gopsutil`. If you are building for the first time:
```bash
go mod tidy
```

## Usage

1. **Run the application**:
   ```bash
   go run main.go
   ```
   *Note: To see advanced hardware details like DDR Memory Type, run with root:*
   ```bash
   sudo go run main.go
   ```

2. **Open dashboard**:
   Visit `http://localhost:8993` in your browser.

## API

You can access the raw metrics data JSON endpoint:
- **URL**: `http://localhost:8993/api.json`
- **Method**: `GET`
- **Response**: Full `SystemStats` object including hardware info and process list.

## License
MIT
