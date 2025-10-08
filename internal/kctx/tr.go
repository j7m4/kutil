package kctx

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ContextChange struct {
	OldName string
	NewName string
}

func ReplaceInContexts(inputRegex, replacement string, force bool) error {
	re, err := regexp.Compile(inputRegex)
	if err != nil {
		return fmt.Errorf("error compiling regex '%s': %v", inputRegex, err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("error loading kubeconfig: %v", err)
	}

	var changes []ContextChange
	for contextName := range rawConfig.Contexts {
		if re.MatchString(contextName) {
			newName := re.ReplaceAllString(contextName, replacement)
			if newName != contextName {
				changes = append(changes, ContextChange{OldName: contextName, NewName: newName})
			}
		}
	}

	if len(changes) == 0 {
		fmt.Printf("No contexts matched regex: %s\n", inputRegex)
		return nil
	}

	for _, change := range changes {
		fmt.Printf("%s -> %s\n", change.OldName, change.NewName)
	}

	if !force && !confirmChanges(len(changes)) {
		fmt.Println("Operation cancelled.")
		return nil
	}

	return applyContextChanges(&rawConfig, changes, loadingRules.GetDefaultFilename())
}

func DeleteFromContexts(deletionRegex string, force bool) error {
	re, err := regexp.Compile(deletionRegex)
	if err != nil {
		return fmt.Errorf("error compiling regex '%s': %v", deletionRegex, err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("error loading kubeconfig: %v", err)
	}

	var changes []ContextChange
	for contextName := range rawConfig.Contexts {
		if re.MatchString(contextName) {
			newName := re.ReplaceAllString(contextName, "")
			if newName != contextName {
				changes = append(changes, ContextChange{OldName: contextName, NewName: newName})
			}
		}
	}

	if len(changes) == 0 {
		fmt.Printf("No contexts matched regex: %s\n", deletionRegex)
		return nil
	}

	for _, change := range changes {
		fmt.Printf("%s -> %s\n", change.OldName, change.NewName)
	}

	if !force && !confirmChanges(len(changes)) {
		fmt.Println("Operation cancelled.")
		return nil
	}

	return applyContextChanges(&rawConfig, changes, loadingRules.GetDefaultFilename())
}

func confirmChanges(count int) bool {
	fmt.Printf("Apply changes to %d context(s)? [y/N]: ", count)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func applyContextChanges(config *clientcmdapi.Config, changes []ContextChange, configPath string) error {
	for _, change := range changes {
		if context, exists := config.Contexts[change.OldName]; exists {
			config.Contexts[change.NewName] = context
			delete(config.Contexts, change.OldName)

			if config.CurrentContext == change.OldName {
				config.CurrentContext = change.NewName
			}
		}
	}

	if err := clientcmd.WriteToFile(*config, configPath); err != nil {
		return fmt.Errorf("error writing kubeconfig: %v", err)
	}

	fmt.Printf("Successfully renamed %d context(s)\n", len(changes))
	return nil
}
