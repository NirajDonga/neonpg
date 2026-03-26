package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/NirajDonga/dbpods/internal/config"
	"github.com/NirajDonga/dbpods/internal/handler"
	"github.com/NirajDonga/dbpods/internal/k8s"
	"github.com/NirajDonga/dbpods/internal/proxy"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	cfg := config.LoadConfig()
	clientset, err := initKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}
	log.Println("Successfully connected to Kubernetes cluster!")

	k8sProvisioner := k8s.NewProvisioner(clientset, cfg.Namespace)
	apiHandler := handler.NewAPIHandler(k8sProvisioner)

	http.HandleFunc("/create-database", apiHandler.HandleCreateDB)

	tcpProxy := proxy.NewPostgresProxy(cfg.Namespace)
	go func() {
		// Standard Postgres port is 5432
		if err := tcpProxy.Start("5432"); err != nil {
			log.Fatalf("TCP Proxy crashed: %v", err)
		}
	}()

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Control Plane API running on %s (Namespace: %s)", serverAddr, cfg.Namespace)

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func initKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return nil, homeErr
		}

		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
