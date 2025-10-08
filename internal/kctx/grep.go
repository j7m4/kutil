package kctx

import (
	"fmt"
	"regexp"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
)

func GrepContexts(regex string, invertMatch bool) error {
	re, err := regexp.Compile(regex)
	if err != nil {
		return fmt.Errorf("error compiling regex '%s': %v", regex, err)
	}

	contexts, err := getContexts()
	if err != nil {
		return err
	}

	var matchedContexts []string
	for _, contextName := range contexts {
		matches := re.MatchString(contextName)
		if (invertMatch && !matches) || (!invertMatch && matches) {
			matchedContexts = append(matchedContexts, contextName)
		}
	}

	sort.Strings(matchedContexts)
	for _, contextName := range matchedContexts {
		fmt.Println(contextName)
	}

	hasMatches := len(matchedContexts) > 0

	if !hasMatches {
		if invertMatch {
			fmt.Printf("All contexts matched regex: %s\n", regex)
		} else {
			fmt.Printf("No contexts matched regex: %s\n", regex)
		}
	}

	return nil
}

func getContexts() ([]string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %v", err)
	}

	var contexts []string
	for contextName := range rawConfig.Contexts {
		contexts = append(contexts, contextName)
	}

	return contexts, nil
}
