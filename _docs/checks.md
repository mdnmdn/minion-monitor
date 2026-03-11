# Monitoring Checks

This tool performs several checks on both servers (via SSH) and web applications (via HTTP/HTTPS).

## Server Checks (via SSH)

- **OS Distribution**: Reports the pretty name of the operating system (from `/etc/os-release`).
- **Kernel Version**: Reports the kernel version (`uname -r`).
- **Logged in Users**: Reports the number of users currently logged in via SSH or local sessions.
- **Updates (Apt)**: For Ubuntu/Debian, it reports the number of pending packages that need update, specifically highlighting security updates.
  - **Critical Alert**: If there are **security updates** pending (>0), the tool marks the report with an error state to trigger notifications.
- **Disk Space (Blocks)**: Checks the disk usage on the root partition (`/`). It reports used vs total blocks and percentage.
- **Memory Used**: Reports used vs total memory in megabytes and percentage.
- **CPU Used**: Calculates CPU usage based on the `top` command (100% - idle%).
- **Load Average**: Reports the 1, 5, and 15-minute load averages from `/proc/loadavg`.
- **Uptime**: Reports how long the server has been running using `uptime -p`.
- **Clock Alignment**: Compares the remote server's Unix timestamp with the local time.
  - **Warning**: If the time drift is greater than **60 seconds**, it triggers an error state.
- **Docker Status**: If enabled in config, it lists the names, statuses, and uptime (how long it has been running) of Docker containers.
- **Top Processes**: (In verbose mode `-v`) Shows the top 10 most CPU-intensive processes.

## Web Application Checks (via HTTP/HTTPS)

- **Status**: Performs an HTTP GET request and checks if the status code is between 200 and 399.
- **SSL Certificate**:
  - Checks if the certificate is present.
  - Validates the expiration date.
  - **Warning**: If the certificate expires in less than 15 days, it flags a warning.
  - Supports `ignore-certificate: true` for self-signed or internal sites.
