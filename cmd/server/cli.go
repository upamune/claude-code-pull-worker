package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

)

type CLI struct {
	Server Server `cmd:"" help:"Run the webhook server (default)" default:"1"`
	SystemdInstall SystemdInstall `cmd:"" help:"Generate systemd service file"`
}

type Server struct {
	ConfigFile string `help:"Path to config file" env:"CONFIG_FILE"`
}

type SystemdInstall struct {
	User       string `help:"User to run the service as" required:""`
	WorkingDir string `help:"Working directory for the service" type:"path" default:"."`
	BinaryPath string `help:"Path to the claude-code-pull-worker binary" type:"path" default:"./claude-code-pull-worker"`
	EnvFile    string `help:"Path to environment file" type:"path" default:".env"`
	Output     string `help:"Output file path" type:"path" default:"claude-code-pull-worker.service"`
}

const systemdTemplate = `[Unit]
Description=Claude Code Pull Worker
After=network.target

[Service]
Type=simple
User={{.User}}
WorkingDirectory={{.WorkingDir}}
ExecStart={{.BinaryPath}} server
Restart=always
RestartSec=5

# Environment variables
EnvironmentFile={{.EnvFile}}
Environment="PATH=/usr/local/bin:/usr/bin:/bin:{{.Home}}/.local/bin"

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths={{.WorkingDir}}

[Install]
WantedBy=multi-user.target
`

func (s *SystemdInstall) Run() error {
	// Convert paths to absolute
	workingDir, err := filepath.Abs(s.WorkingDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for working directory: %w", err)
	}

	binaryPath, err := filepath.Abs(s.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for binary: %w", err)
	}

	envFile, err := filepath.Abs(s.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for env file: %w", err)
	}

	// Get home directory for the user
	home := fmt.Sprintf("/home/%s", s.User)
	if s.User == "root" {
		home = "/root"
	}

	// Parse and execute template
	tmpl, err := template.New("systemd").Parse(systemdTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	output, err := os.Create(s.Output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	data := struct {
		User       string
		WorkingDir string
		BinaryPath string
		EnvFile    string
		Home       string
	}{
		User:       s.User,
		WorkingDir: workingDir,
		BinaryPath: binaryPath,
		EnvFile:    envFile,
		Home:       home,
	}

	if err := tmpl.Execute(output, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Systemd service file generated: %s\n", s.Output)
	fmt.Println("\nTo install the service:")
	fmt.Printf("  sudo cp %s /etc/systemd/system/\n", s.Output)
	fmt.Println("  sudo systemctl daemon-reload")
	fmt.Println("  sudo systemctl enable claude-code-pull-worker")
	fmt.Println("  sudo systemctl start claude-code-pull-worker")

	return nil
}