package awsctx

import (
	"fmt"
	"os"
)

func handleRegion(args []string) error {
	if len(args) == 0 {
		if isInteractive() && hasFzf() {
			return chooseRegionInteractive()
		}
		return listRegions()
	}

	switch args[0] {
	case "-c", "--current":
		showCurrentRegion()
		return nil
	case "-":
		return swapRegion()
	case "-h", "--help":
		printRegionUsage()
		return nil
	default:
		return setRegion(args[0])
	}
}

func listRegions() error {
	cur := currentRegion()
	for _, r := range awsRegions {
		if r == cur {
			fmt.Fprintf(os.Stderr, "\033[33m\033[40m%s\033[0m\n", r)
		} else {
			fmt.Fprintln(os.Stderr, r)
		}
	}
	return nil
}

func showCurrentRegion() {
	fmt.Fprintln(os.Stderr, currentRegion())
}

func setRegion(name string) error {
	if !isValidRegion(name) {
		return fmt.Errorf("unknown AWS region: %s", name)
	}

	prev := currentRegion()
	if prev != name && prev != "(none)" {
		savePrevious("region", prev)
	}

	if err := switchRegionInConfig(name); err != nil {
		return err
	}

	saveState("region", name)

	fmt.Fprintf(os.Stderr, "Switched to region: %s\n", name)
	return nil
}

func swapRegion() error {
	prev := readPrevious("region")
	if prev == "" {
		return fmt.Errorf("no previous region found")
	}
	return setRegion(prev)
}

func chooseRegionInteractive() error {
	choice, err := runFzf("region")
	if err != nil {
		return err
	}
	if choice == "" {
		return fmt.Errorf("no region selected")
	}
	return setRegion(choice)
}

func printRegionUsage() {
	fmt.Fprint(os.Stderr, `USAGE:
  awsctx region              list regions (fzf if available)
  awsctx region <NAME>       switch to region <NAME>
  awsctx region -            switch to previous region
  awsctx region -c           show current region
`)
}
