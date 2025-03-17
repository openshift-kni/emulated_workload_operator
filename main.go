// Package main ...
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	emulated_workload "github.com/openshift-kni/emulated_workload_operator/internal/workload"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	kube_rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	var kubeConfig *kube_rest.Config
	var err error

	// read config from env
	// if not in cluster
	kubeCfgPath := os.Getenv("KUBECONFIG")

	if kubeCfgPath == "" {
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Error getting in-cluster config: %v\n", err)
			log.Println("Error getting in-cluster config: %v\n", err)
			os.Exit(1)
		}
	} else {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeCfgPath)

		if err != nil {
			fmt.Printf("Error getting kubeconfig from env: %v\n", err)
			log.Println("Error getting kubeconfig from env: %v\n", err)
			os.Exit(1)
		}

	}

	// read env for the workload  config file
	workloadCfgPath := os.Getenv("WORKLOADPATH")

	if workloadCfgPath == "" {
		if err != nil {
			fmt.Printf("Error getting workload path env: %v\n", err)
			log.Println("Error getting workload path env:: %v\n", err)
			os.Exit(1)
		}
		//env not set use default value
		workloadCfgPath = "/operand-values/workload.yaml"
	}
	//decode the workload file for

	// Create the clientset
	clientSet, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("Error creating clientset: %v\n", err)
		log.Println("Error creating clientset: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// check pod in the test space
	//pods, err := clientset.CoreV1().Pods("test").Get(context.TODO(), "workload", metav1.GetOptions{})
	result := emulated_workload.DeleteWorkloadPod(ctx, clientSet, true)

	if result == false {
		fmt.Printf("failed to delete existing workload pod")
		log.Println("failed to delete existing workload pod")
		os.Exit(1)
	}

	result = emulated_workload.ApplyWorkloadPod(ctx, clientSet, workloadCfgPath)
	if result == false {
		fmt.Printf("failed to apply existing workload pod")
		log.Println("failed to apply existing workload pod")
	}

	//sleep
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(60 * time.Second)
		}
	}()

	<-cancelChan
	log.Println("Recieved signal handling exit")
	emulated_workload.DeleteWorkloadPod(ctx, clientSet, false)

	wg.Wait()
}
