package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type FullReport struct {
	Project   string          `json:"project" yaml:"project"`
	Time      string          `json:"time" yaml:"time"`
	Servers   []*ServerReport `json:"servers" yaml:"servers"`
	HasErrors bool            `json:"has_errors" yaml:"has_errors"`
}

var (
	configPath string
	verbose    bool
	format     string
	hardFail   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "minion-mon",
		Short: "A simple server and webapp monitoring tool",
		Long:  `A Go CLI tool that generates reports about servers (via SSH) and web applications (via HTTP).`,
		Run: func(cmd *cobra.Command, args []string) {
			runMonitor()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "hosts.yaml", "Path to config file (yaml or toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode (show top processes)")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "text", "Output format: text, markdown, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&hardFail, "hard-fail", false, "Exit with non-zero code if any error is detected")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runMonitor() {
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fullReport := &FullReport{
		Project: config.Project.Name,
		Time:    time.Now().Format(time.RFC1123),
	}

	for name, server := range config.Servers {
		if format == "text" {
			fmt.Printf("Checking server: %s (%s)...\n", name, server.Host)
		}
		report, err := CheckServer(name, server, verbose)
		if err != nil {
			// Mock a report with error info
			errReport := &ServerReport{Name: name, OS: fmt.Sprintf("ERROR: %v", err)}
			fullReport.Servers = append(fullReport.Servers, errReport)
			fullReport.HasErrors = true
			continue
		}
		fullReport.Servers = append(fullReport.Servers, report)
		for _, wr := range report.WebReports {
			if strings.Contains(wr.Status, "DOWN") || wr.CertWarning {
				fullReport.HasErrors = true
			}
		}
		if report.SecUpdates != "0" && report.SecUpdates != "" {
			fullReport.HasErrors = true
		}
		if report.TimeDrift > 60 || report.TimeDrift < -60 {
			fullReport.HasErrors = true
		}
	}

	var output string
	switch format {
	case "json":
		data, _ := json.MarshalIndent(fullReport, "", "  ")
		output = string(data)
	case "yaml":
		data, _ := yaml.Marshal(fullReport)
		output = string(data)
	case "markdown":
		output = generateMarkdown(fullReport)
	default:
		output = generateText(fullReport, verbose)
	}

	fmt.Println(output)

	// Telegram Notification
	if config.Alert.Telegram.Enabled {
		shouldSend := false
		if config.Alert.Telegram.Mode == "always" {
			shouldSend = true
		} else if config.Alert.Telegram.Mode == "error" && fullReport.HasErrors {
			shouldSend = true
		}

		if shouldSend {
			if format == "text" {
				fmt.Println("Sending Telegram notification...")
			}
			SendTelegram(config.Alert.Telegram.Token, config.Alert.Telegram.ChatID, generateText(fullReport, verbose))
		}
	}

	if hardFail && fullReport.HasErrors {
		os.Exit(1)
	}
}

func generateText(r *FullReport, verbose bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Project: %s\n", r.Project))
	b.WriteString(fmt.Sprintf("Time: %s\n", r.Time))
	b.WriteString(strings.Repeat("-", 40) + "\n")

	for _, s := range r.Servers {
		if strings.HasPrefix(s.OS, "ERROR:") {
			b.WriteString(fmt.Sprintf("[-] Server %s: %s\n", s.Name, s.OS))
			continue
		}
		b.WriteString(fmt.Sprintf("[+] Server: %s\n", s.Name))
		b.WriteString(fmt.Sprintf("    OS:     %s (%s)\n", s.OS, s.Kernel))
		b.WriteString(fmt.Sprintf("    Users:  %s logged in\n", s.Users))
		if s.Updates != "" {
			b.WriteString(fmt.Sprintf("    Updates: %s total (%s security)\n", s.Updates, s.SecUpdates))
		}
		b.WriteString(fmt.Sprintf("    Disk:   %s\n", s.DiskSpace))
		b.WriteString(fmt.Sprintf("    Mem:    %s\n", s.MemoryUsed))
		b.WriteString(fmt.Sprintf("    CPU:    %s\n", s.CPUUsed))
		b.WriteString(fmt.Sprintf("    Load:   %s\n", s.LoadAverage))
		b.WriteString(fmt.Sprintf("    Uptime: %s\n", s.Uptime))

		driftLabel := fmt.Sprintf("%ds drift", s.TimeDrift)
		if s.TimeDrift > 60 || s.TimeDrift < -60 {
			b.WriteString(fmt.Sprintf("    Clock:  WARNING (%s)\n", driftLabel))
		} else {
			b.WriteString(fmt.Sprintf("    Clock:  OK (%s)\n", driftLabel))
		}

		if s.SARStats != "" {
			b.WriteString(fmt.Sprintf("    History: %s\n", s.SARStats))
		}
		if s.DockerStatus != "" {
			b.WriteString(fmt.Sprintf("    Docker: %s\n", s.DockerStatus))
		}
		if verbose && s.TopProcesses != "" {
			b.WriteString("    Top Processes:\n")
			for _, line := range strings.Split(s.TopProcesses, "\n") {
				if line != "" {
					b.WriteString("      " + line + "\n")
				}
			}
		}

		for _, wr := range s.WebReports {
			symbol := "[+]"
			if strings.Contains(wr.Status, "DOWN") {
				symbol = "[-]"
			}
			b.WriteString(fmt.Sprintf("    Webapp %s %s: %s\n", symbol, wr.Name, wr.Status))
			if wr.CertOk {
				certStatus := "OK"
				if wr.CertWarning {
					certStatus = fmt.Sprintf("WARNING (Expires in %v)", time.Until(wr.CertExpiry).Round(time.Hour))
				}
				b.WriteString(fmt.Sprintf("      SSL Cert: %s (Expires: %s)\n", certStatus, wr.CertExpiry.Format("2006-01-02")))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func generateMarkdown(r *FullReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Report for %s\n\n", r.Project))
	b.WriteString(fmt.Sprintf("*Generated at: %s*\n\n", r.Time))

	for _, s := range r.Servers {
		b.WriteString(fmt.Sprintf("## Server: %s\n", s.Name))
		if strings.HasPrefix(s.OS, "ERROR:") {
			b.WriteString(fmt.Sprintf("> **ERROR**: %s\n\n", s.OS))
			continue
		}
		b.WriteString(fmt.Sprintf("- **OS**: %s (%s)\n", s.OS, s.Kernel))
		b.WriteString(fmt.Sprintf("- **Users**: %s\n", s.Users))
		if s.Updates != "" {
			b.WriteString(fmt.Sprintf("- **Updates**: %s (%s security)\n", s.Updates, s.SecUpdates))
		}
		b.WriteString(fmt.Sprintf("- **Disk**: %s\n", s.DiskSpace))
		b.WriteString(fmt.Sprintf("- **Memory**: %s\n", s.MemoryUsed))
		b.WriteString(fmt.Sprintf("- **CPU**: %s\n", s.CPUUsed))
		b.WriteString(fmt.Sprintf("- **Load**: %s\n", s.LoadAverage))
		b.WriteString(fmt.Sprintf("- **Uptime**: %s\n", s.Uptime))
		b.WriteString(fmt.Sprintf("- **Clock**: %ds drift\n", s.TimeDrift))

		if s.SARStats != "" {
			b.WriteString(fmt.Sprintf("- **History**: %s\n", s.SARStats))
		}
		if s.DockerStatus != "" {
			b.WriteString(fmt.Sprintf("- **Docker**: `%s`\n", s.DockerStatus))
		}

		if len(s.WebReports) > 0 {
			b.WriteString("\n### Web Applications\n\n")
			b.WriteString("| Name | URL | Status | SSL Cert |\n")
			b.WriteString("|------|-----|--------|----------|\n")
			for _, wr := range s.WebReports {
				certInfo := "N/A"
				if wr.CertOk {
					certInfo = wr.CertExpiry.Format("2006-01-02")
					if wr.CertWarning {
						certInfo = "⚠️ " + certInfo
					}
				}
				status := wr.Status
				if strings.Contains(status, "DOWN") {
					status = "❌ " + status
				} else {
					status = "✅ " + status
				}
				b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", wr.Name, wr.URL, status, certInfo))
			}
		}
		b.WriteString("\n---\n")
	}
	return b.String()
}
