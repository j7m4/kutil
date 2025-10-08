package kctx

import (
	"fmt"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
)

func ListContexts() error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("error loading kubeconfig: %v", err)
	}

	if len(rawConfig.Contexts) == 0 {
		fmt.Println("No contexts found in kubeconfig")
		return nil
	}

	if rawConfig.CurrentContext != "" {
		fmt.Printf("* %s\n", rawConfig.CurrentContext)
	}

	var otherContexts []string
	for contextName := range rawConfig.Contexts {
		if contextName != rawConfig.CurrentContext {
			otherContexts = append(otherContexts, contextName)
		}
	}

	sort.Strings(otherContexts)
	for _, contextName := range otherContexts {
		fmt.Printf("  %s\n", contextName)
	}

	return nil
}
