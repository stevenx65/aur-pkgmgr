package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func detectAURHelper() string {
	for _, name := range []string{"paru", "yay"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return ""
}

func searchPackages(query string, aurHelper string) ([]Package, error) {
	if aurHelper != "" {
		return searchWith(aurHelper, query)
	}
	return searchWith("pacman", query)
}

func searchWith(cmd string, query string) ([]Package, error) {
	args := []string{"-Ss", query}
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " ") {
			if len(pkgs) > 0 {
				pkgs[len(pkgs)-1].Description = strings.TrimSpace(line)
			}
			continue
		}
		pkg := parsePkgLine(line)
		if pkg.Name != "" {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs, nil
}

func parsePkgLine(line string) Package {
	// format: repo/pkgname version (groups)
	// or: aur/pkgname version
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return Package{}
	}

	repoPkg := parts[0]
	repo := ""
	pkgName := ""

	if idx := strings.Index(repoPkg, "/"); idx != -1 {
		repo = repoPkg[:idx]
		pkgName = repoPkg[idx+1:]
	} else {
		pkgName = repoPkg
	}

	version := parts[1]

	return Package{
		Name:       pkgName,
		Version:    version,
		Repository: repo,
	}
}

func listInstalled() ([]Package, error) {
	out, err := exec.Command("pacman", "-Qs").Output()
	if err != nil {
		return nil, fmt.Errorf("list installed failed: %w", err)
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var currentPkg *Package
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " ") {
			if currentPkg != nil {
				currentPkg.Description = strings.TrimSpace(line)
			}
			continue
		}
		// local/pkgname version
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		repoPkg := parts[0]
		pkgName := ""
		if idx := strings.Index(repoPkg, "/"); idx != -1 {
			pkgName = repoPkg[idx+1:]
		} else {
			pkgName = repoPkg
		}
		version := strings.TrimPrefix(parts[1], "(")
		version = strings.TrimSuffix(version, ")")

		pkg := Package{
			Name:             pkgName,
			Version:          version,
			InstalledVersion: version,
			Repository:       "",
		}
		pkgs = append(pkgs, pkg)
		currentPkg = &pkgs[len(pkgs)-1]
	}
	return pkgs, nil
}

func checkUpdates(aurHelper string) ([]Package, error) {
	// Try with AUR helper first
	if aurHelper != "" {
		out, err := exec.Command(aurHelper, "-Qu").Output()
		if err == nil {
			return parseUpdateOutput(string(out)), nil
		}
	}
	// Fallback to pacman
	out, err := exec.Command("pacman", "-Qu").Output()
	if err != nil {
		return nil, nil // no updates or error
	}
	return parseUpdateOutput(string(out)), nil
}

func parseUpdateOutput(out string) []Package {
	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// format: pkgname oldver -> newver
		// or: repo/pkgname oldver -> newver (aur helper)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pkgName := parts[0]
		if idx := strings.Index(pkgName, "/"); idx != -1 {
			pkgName = pkgName[idx+1:]
		}
		pkg := Package{
			Name:    pkgName,
			Version: parts[len(parts)-1],
		}
		// If there's an arrow pattern, old version is before it
		if len(parts) >= 4 && parts[len(parts)-2] == "->" {
			pkg.InstalledVersion = parts[1]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func getPackageInfo(pkg Package, aurHelper string) (string, error) {
	if pkg.Repository == "aur" || (pkg.InstalledVersion == "" && aurHelper != "") {
		return getInfo(aurHelper, pkg.Name)
	}
	// Try pacman -Qi first (installed), then -Si (available)
	out, err := exec.Command("pacman", "-Qi", pkg.Name).Output()
	if err == nil && len(out) > 0 {
		return formatInfo(string(out), false), nil
	}
	return getInfo("pacman", pkg.Name)
}

func getInfo(cmd, pkgName string) (string, error) {
	out, err := exec.Command(cmd, "-Si", pkgName).Output()
	if err != nil {
		return "", fmt.Errorf("info failed: %w", err)
	}
	return formatInfo(string(out), cmd != "pacman"), nil
}

func formatInfo(out string, aur bool) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(out))
	fields := []string{
		"Name", "Version", "Description", "Repository",
		"URL", "Licenses", "Groups", "Depends On",
		"Optional Deps", "Required By", "Conflicts With",
		"Installed Size", "Packager", "Build Date",
	}
	for scanner.Scan() {
		line := scanner.Text()
		for _, field := range fields {
			if strings.HasPrefix(line, field+" :") || strings.HasPrefix(line, field+":") {
				sb.WriteString(line)
				sb.WriteString("\n")
				break
			}
		}
	}
	return sb.String()
}

func execCmd(args []string) (string, error) {
	cmd := exec.Command("sudo", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func installPkg(pkg Package, aurHelper string) (string, error) {
	if aurHelper != "" && pkg.Repository == "aur" {
		out, err := exec.Command(aurHelper, "-S", "--noconfirm", pkg.Name).CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return string(out), nil
	}
	return execCmd([]string{"pacman", "-S", "--noconfirm", pkg.Name})
}

func removePkg(pkg Package) (string, error) {
	return execCmd([]string{"pacman", "-Rns", "--noconfirm", pkg.Name})
}

func updatePkg(pkg Package, aurHelper string) (string, error) {
	if aurHelper != "" {
		out, err := exec.Command(aurHelper, "-S", "--noconfirm", pkg.Name).CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return string(out), nil
	}
	return execCmd([]string{"pacman", "-S", "--noconfirm", pkg.Name})
}

func systemUpgrade(aurHelper string) (string, error) {
	if aurHelper != "" {
		out, err := exec.Command(aurHelper, "-Syu", "--noconfirm").CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return string(out), nil
	}
	return execCmd([]string{"pacman", "-Syu", "--noconfirm"})
}

func detectTerminal() string {
	if t := os.Getenv("TERMINAL"); t != "" {
		if _, err := exec.LookPath(t); err == nil {
			return t
		}
	}
	for _, t := range []string{"alacritty", "kitty", "foot", "wezterm", "gnome-terminal", "konsole", "xfce4-terminal", "xterm"} {
		if _, err := exec.LookPath(t); err == nil {
			return t
		}
	}
	return "xterm"
}

func buildTerminalCmd(cmdLine string) *exec.Cmd {
	term := detectTerminal()
	pause := `; echo ""; echo "━━ Press Enter to return to PkgMgr ━━"; read`
	fullCmd := cmdLine + pause

	switch term {
	case "alacritty", "kitty", "foot":
		return exec.Command(term, "-e", "sh", "-c", fullCmd)
	case "gnome-terminal":
		return exec.Command(term, "--", "sh", "-c", fullCmd)
	case "konsole":
		return exec.Command(term, "-e", "sh", "-c", fullCmd)
	case "xfce4-terminal":
		return exec.Command(term, "-e", "sh", "-c", fullCmd)
	case "wezterm":
		return exec.Command(term, "start", "--", "sh", "-c", fullCmd)
	default:
		return exec.Command(term, "-e", "sh", "-c", fullCmd)
	}
}
