package export

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// PandocWrapper handles Pandoc command execution
type PandocWrapper struct {
	timeout time.Duration
}

// NewPandocWrapper creates a new Pandoc wrapper
func NewPandocWrapper() *PandocWrapper {
	return &PandocWrapper{
		timeout: 30 * time.Second,
	}
}

// CheckPandocInstalled checks if Pandoc is available
func (p *PandocWrapper) CheckPandocInstalled() error {
	cmd := exec.Command("pandoc", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc not found: please install pandoc to enable document export")
	}
	return nil
}

// ConvertMarkdownToFormat converts markdown to specified format using Pandoc
func (p *PandocWrapper) ConvertMarkdownToFormat(markdownContent string, outputPath string, format string) error {
	// Check if Pandoc is installed
	if err := p.CheckPandocInstalled(); err != nil {
		return err
	}

	// Build Pandoc command arguments
	args := []string{
		"-f", "markdown",
		"-o", outputPath,
	}

	// Add format-specific options
	switch format {
	case "pdf":
		// Use xelatex for Unicode support (emojis), images should work now
		args = append(args, "--pdf-engine=xelatex")
	case "docx":
		// DOCX needs no special options
	case "html":
		// Standalone HTML with embedded CSS
		args = append(args, "--standalone")
		args = append(args, "--self-contained")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Create command
	cmd := exec.Command("pandoc", args...)

	// Provide markdown content via stdin
	cmd.Stdin = bytes.NewBufferString(markdownContent)

	// Capture stderr for error messages
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("pandoc conversion failed: %s", stderr.String())
			}
			return fmt.Errorf("pandoc conversion failed: %w", err)
		}
		return nil
	case <-time.After(p.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("pandoc conversion timed out after %v", p.timeout)
	}
}

// GetPandocVersion returns the installed Pandoc version
func (p *PandocWrapper) GetPandocVersion() (string, error) {
	cmd := exec.Command("pandoc", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pandoc version: %w", err)
	}
	
	// Return first line which contains version
	lines := bytes.Split(output, []byte("\n"))
	if len(lines) > 0 {
		return string(lines[0]), nil
	}
	return "unknown", nil
}

// ConvertMarkdownToFormatInDir converts markdown to specified format using Pandoc from a specific working directory
func (p *PandocWrapper) ConvertMarkdownToFormatInDir(markdownContent string, outputPath string, format string, workingDir string) error {
	// Check if Pandoc is installed
	if err := p.CheckPandocInstalled(); err != nil {
		return err
	}

	// Build Pandoc command arguments
	args := []string{
		"-f", "markdown",
		"-o", outputPath,
	}

	// Add format-specific options
	switch format {
	case "pdf":
		// Use xelatex for Unicode support (emojis), images should work now
		args = append(args, "--pdf-engine=xelatex")
	case "docx":
		// DOCX needs no special options
	case "html":
		// Standalone HTML with embedded CSS
		args = append(args, "--standalone")
		args = append(args, "--self-contained")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Create command with working directory
	cmd := exec.Command("pandoc", args...)
	cmd.Dir = workingDir

	// Provide markdown content via stdin
	cmd.Stdin = bytes.NewBufferString(markdownContent)

	// Capture stderr for error messages
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("pandoc conversion failed: %s", stderr.String())
			}
			return fmt.Errorf("pandoc conversion failed: %w", err)
		}
		return nil
	case <-time.After(p.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("pandoc conversion timed out after %v", p.timeout)
	}
}