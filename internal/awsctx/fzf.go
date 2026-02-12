package awsctx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func isInteractive() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// runFzf launches fzf for interactive selection.
// subcommand is "profile" or "region".
// Returns the user's choice or empty string if cancelled.
func runFzf(subcommand string) (string, error) {
	selfCmd, _ := os.Executable()
	if selfCmd == "" {
		selfCmd = os.Args[0]
	}

	cmd := exec.Command("fzf", "--ansi", "--no-preview")
	var out bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = &out
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("FZF_DEFAULT_COMMAND=%s --fzf-list %s", selfCmd, subcommand),
		"_AWSCTX_FORCE_COLOR=1",
	)

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// fzf was cancelled (e.g. Esc/Ctrl-C)
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

// fzfList prints items to stdout for fzf consumption.
func fzfList(subcommand string) error {
	switch subcommand {
	case "profile":
		profiles, err := getProfiles()
		if err != nil {
			return err
		}
		cur := currentProfile()
		forceColor := os.Getenv("_AWSCTX_FORCE_COLOR") == "1"
		for _, p := range profiles {
			if forceColor && p == cur {
				fmt.Printf("\033[33m\033[40m%s\033[0m\n", p)
			} else {
				fmt.Println(p)
			}
		}
	case "region":
		cur := currentRegion()
		forceColor := os.Getenv("_AWSCTX_FORCE_COLOR") == "1"
		for _, r := range awsRegions {
			if forceColor && r == cur {
				fmt.Printf("\033[33m\033[40m%s\033[0m\n", r)
			} else {
				fmt.Println(r)
			}
		}
	default:
		return fmt.Errorf("unknown subcommand for --fzf-list: %s", subcommand)
	}
	return nil
}
