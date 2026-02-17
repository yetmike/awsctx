package awsctx

import (
	"fmt"
	"os"
	"os/exec"
)

var Version = "v0.0.1"

func Run(args []string) error {
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("aws CLI is not installed. Install it from https://aws.amazon.com/cli/")
	}

	if len(args) < 2 {
		return ShowStatus()
	}

	switch args[1] {
	case "profile", "p":
		return handleProfile(args[2:])
	case "region", "r":
		return handleRegion(args[2:])
	case "--fzf-list":
		if len(args) < 3 {
			return fmt.Errorf("missing subcommand for --fzf-list")
		}
		return fzfList(args[2])
	case "-h", "--help":
		printUsage()
		return nil
	case "-v", "--version":
		fmt.Fprintf(os.Stderr, "awsctx %s\n", Version)
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nRun 'awsctx --help' for usage", args[1])
	}
}

func ShowStatus() error {
	profile := currentProfile()
	region := currentRegion()

	fmt.Fprintf(os.Stderr, "profile: %s\n", profile)
	fmt.Fprintf(os.Stderr, "region:  %s\n", region)
	return nil
}

func printUsage() {
	fmt.Fprint(os.Stderr, `USAGE:
  awsctx                          show current profile and region
  awsctx p,  profile [<name>]     list or switch AWS profiles
  awsctx r,  region  [<name>]     list or switch AWS regions

  awsctx <subcommand> -c          show current value
  awsctx <subcommand> -           switch to previous value

  awsctx -h, --help               show this message
  awsctx -v, --version            show version

Switches the [default] profile in ~/.aws/config and ~/.aws/credentials.
Original [default] is backed up and restored with 'awsctx p default'.

COMPLETIONS (optional):
  source /path/to/awsctx/shell/awsctx.sh
`)
}
