package kctx

import (
	"fmt"
	"io"
	"os"
	"time"

	"k8s.io/client-go/tools/clientcmd"
)

func BackupKubeconfig() error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfigPath := loadingRules.GetDefaultFilename()

	if kubeconfigPath == "" {
		return fmt.Errorf("could not determine kubeconfig file path")
	}

	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file does not exist at %s", kubeconfigPath)
	}

	timestamp := time.Now().Format("200601021504")
	backupPath := kubeconfigPath + "_backup_" + timestamp

	sourceFile, err := os.Open(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("error opening kubeconfig file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("error creating backup file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	fmt.Printf("Backup created: %s\n", backupPath)
	return nil
}
