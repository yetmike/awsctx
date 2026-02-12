package awsctx

import (
	"fmt"
	"os"
)

func handleProfile(args []string) error {
	if len(args) == 0 {
		if isInteractive() && hasFzf() {
			return chooseProfileInteractive()
		}
		return listProfiles()
	}

	switch args[0] {
	case "-c", "--current":
		showCurrentProfile()
		return nil
	case "-":
		return swapProfile()
	case "-h", "--help":
		printProfileUsage()
		return nil
	default:
		return setProfile(args[0])
	}
}

func listProfiles() error {
	profiles, err := getProfiles()
	if err != nil {
		return err
	}
	cur := currentProfile()
	for _, p := range profiles {
		if p == cur {
			fmt.Fprintf(os.Stderr, "\033[33m\033[40m%s\033[0m\n", p)
		} else {
			fmt.Fprintln(os.Stderr, p)
		}
	}
	return nil
}

func showCurrentProfile() {
	fmt.Fprintln(os.Stderr, currentProfile())
}

func setProfile(name string) error {
	if !profileExists(name) {
		return fmt.Errorf("profile %q not found in %s", name, awsConfigPath())
	}

	prev := currentProfile()
	if prev != name {
		savePrevious("profile", prev)
	}

	if err := switchProfileInConfig(name); err != nil {
		return err
	}
	if err := switchProfileInCredentials(name); err != nil {
		return err
	}

	saveState("profile", name)

	fmt.Fprintf(os.Stderr, "Switched to profile: %s\n", name)
	return nil
}

func swapProfile() error {
	prev := readPrevious("profile")
	if prev == "" {
		return fmt.Errorf("no previous profile found")
	}
	return setProfile(prev)
}

func chooseProfileInteractive() error {
	choice, err := runFzf("profile")
	if err != nil {
		return err
	}
	if choice == "" {
		return fmt.Errorf("no profile selected")
	}
	return setProfile(choice)
}

func printProfileUsage() {
	fmt.Fprint(os.Stderr, `USAGE:
  awsctx profile              list profiles (fzf if available)
  awsctx profile <NAME>       switch to profile <NAME>
  awsctx profile -            switch to previous profile
  awsctx profile -c           show current profile
`)
}
