package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type ServerReport struct {
	Name         string
	OS           string
	Kernel       string
	Users        string
	Updates      string
	SecUpdates   string
	DiskSpace    string
	MemoryUsed   string
	CPUUsed      string
	LoadAverage  string
	Uptime       string
	TimeDrift    int // seconds
	SARStats     string
	TopProcesses string
	DockerStatus string
	WebReports   []WebReport
}

type WebReport struct {
	Name         string
	URL          string
	Status       string
	CertOk       bool
	CertExpiry   time.Time
	CertWarning  bool
	ErrorMessage string
}

func GetSSHClient(server Server) (*ssh.Client, error) {
	var auth []ssh.AuthMethod

	if server.Credentials.SSHKey != "" {
		keyPath := server.Credentials.SSHKey
		if strings.HasPrefix(keyPath, "~/") {
			home, _ := os.UserHomeDir()
			keyPath = filepath.Join(home, keyPath[2:])
		}
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if server.Credentials.Password != "" {
		auth = append(auth, ssh.Password(server.Credentials.Password))
	}

	config := &ssh.ClientConfig{
		User:            server.Credentials.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return ssh.Dial("tcp", server.Host+":22", config)
}

func RunCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		return "", err
	}
	return b.String(), nil
}

func CheckServer(name string, server Server, verbose bool) (*ServerReport, error) {
	report := &ServerReport{Name: name}

	client, err := GetSSHClient(server)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// OS Distribution
	osDist, _ := RunCommand(client, "cat /etc/os-release | grep PRETTY_NAME | cut -d '\"' -f 2")
	report.OS = strings.TrimSpace(osDist)

	// Kernel
	kernel, _ := RunCommand(client, "uname -r")
	report.Kernel = strings.TrimSpace(kernel)

	// Users
	users, _ := RunCommand(client, "who | wc -l")
	report.Users = strings.TrimSpace(users)

	// Updates (Debian-based)
	if strings.Contains(strings.ToLower(report.OS), "ubuntu") || strings.Contains(strings.ToLower(report.OS), "debian") {
		aptCheck, _ := RunCommand(client, "if [ -f /usr/lib/update-notifier/apt-check ]; then /usr/lib/update-notifier/apt-check 2>&1; else echo '0;0'; fi")
		parts := strings.Split(strings.TrimSpace(aptCheck), ";")
		if len(parts) >= 2 {
			report.Updates = parts[0]
			report.SecUpdates = parts[1]
		}
	}

	// Disk Space (Human Readable)
	disk, _ := RunCommand(client, "df -h / | tail -n 1 | awk '{print $3 \" used of \" $2 \" ( \" $5 \" )\"}'")
	report.DiskSpace = strings.TrimSpace(disk)

	// Mem Used (Human Readable)
	mem, _ := RunCommand(client, "free -h | grep Mem | awk '{print $3 \" / \" $2}'")
	memPct, _ := RunCommand(client, "free | grep Mem | awk '{print int($3/$2 * 100)}'")
	report.MemoryUsed = fmt.Sprintf("%s ( %s%% )", strings.TrimSpace(mem), strings.TrimSpace(memPct))

	// CPU Used
	cpu, _ := RunCommand(client, "top -bn1 | grep 'Cpu(s)' | awk '{print 100 - $8 \"%\"}'")
	report.CPUUsed = strings.TrimSpace(cpu)

	// Load Average
	load, _ := RunCommand(client, "cat /proc/loadavg | awk '{print $1 \" (1m), \" $2 \" (5m), \" $3 \" (15m)\"}'")
	report.LoadAverage = strings.TrimSpace(load)

	// Uptime
	uptime, _ := RunCommand(client, "uptime -p")
	report.Uptime = strings.TrimSpace(uptime)

	// Time Drift
	remoteUnixStr, _ := RunCommand(client, "date +%s")
	var remoteUnix int64
	fmt.Sscanf(strings.TrimSpace(remoteUnixStr), "%d", &remoteUnix)
	if remoteUnix > 0 {
		report.TimeDrift = int(time.Now().Unix() - remoteUnix)
	}

	if server.SAR.Enabled {
		sarPath, _ := RunCommand(client, "command -v sar")
		if strings.TrimSpace(sarPath) != "" {
			// CPU last 24h
			cpuSar, _ := RunCommand(client, "sar -u | grep Average | awk '{print 100-$8 \"%\"}'")
			// Mem last 24h
			memSar, _ := RunCommand(client, "sar -r | grep Average | awk '{print $4 \"%\"}'")
			report.SARStats = fmt.Sprintf("CPU %s, Mem %s (avg last 24h)", strings.TrimSpace(cpuSar), strings.TrimSpace(memSar))
		} else {
			report.SARStats = "sar not installed"
		}
	}

	if verbose {
		top, _ := RunCommand(client, "ps -eo pid,ppid,cmd,%mem,%cpu --sort=-%cpu | head -n 11")
		report.TopProcesses = top
	}

	if server.Docker.Status {
		docker, _ := RunCommand(client, "docker ps --format '{{.Names}} [{{.Status}} - {{.RunningFor}}]' | tr '\n' ','")
		report.DockerStatus = strings.TrimSuffix(docker, ",")
	}

	for webName, webapp := range server.Webapps {
		webReport := CheckWebapp(webName, webapp)
		report.WebReports = append(report.WebReports, webReport)
	}

	return report, nil
}

func CheckWebapp(name string, webapp Webapp) WebReport {
	report := WebReport{Name: name, URL: webapp.URL}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: webapp.IgnoreCertificate},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Get(webapp.URL)
	if err != nil {
		report.Status = "DOWN"
		report.ErrorMessage = err.Error()
		return report
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		report.Status = "UP"
	} else {
		report.Status = fmt.Sprintf("DOWN (%d)", resp.StatusCode)
	}

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		report.CertOk = true
		report.CertExpiry = cert.NotAfter
		if time.Until(cert.NotAfter) < 15*24*time.Hour {
			report.CertWarning = true
		}
	}

	return report
}
