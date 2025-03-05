package workload

import (
	"context"
	"log"
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Config represents the structure of the YAML data
type Config struct {
	Data map[string]string `yaml:"data"`
}

// the function will read a configmap file that contains workload.yaml
// Then generate the workload.yaml in the directory
func PorcessWorkloadCfgfile() bool {
	cmFile, err := os.ReadFile("configmap.yaml")
	if err != nil {
		log.Printf("configmap file does not exist")
		return false
	}

	var cmData Config

	err = yaml.Unmarshal(cmFile, &cmData.Data)
	if err != nil {
		log.Printf("configmap file unmarshal failed")
		return false
	}

	workload, ok := cmData.Data["workload.yaml"]
	if !ok {
		log.Printf("workload.yaml is not found ")
		return false
	}

	//write back workload file to the directory
	err = os.WriteFile("workload.yaml", []byte(workload), 0643)
	if err != nil {
		log.Printf(" write workload.yaml iailed")
		return false
	}

	return true
}

// the function will delete the existing workload pods under test ns
func DeleteWorkloadPod(ctx context.Context, clientSet *kubernetes.Clientset, wait bool) bool {

	podName := "workload"
	ns := "test"

	deletePolicy := metav1.DeletePropagationForeground

	err := clientSet.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil {
		log.Printf(" Delete workload pod failed err=%s \n", err)
	}
	//waiting for the pods deleted
	if wait == false {
		return true
	}

	for i := 0; i < 15; i++ {
		_, err := clientSet.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil && err.Error() == "NotFound" {
			return true
		}
		time.Sleep(time.Duration(1+i*10) * time.Second)
	}

	return false
}

func ApplyWorkloadPod(ctx context.Context, clientSet *kubernetes.Clientset) bool {
	podYaml, err := os.ReadFile("workload.yaml")

	if err != nil {
		log.Println("Failed to read the workload.yaml with err %s, exit", err)
		return false
	}

	podCfg := &corev1.Pod{}

	err = yaml.Unmarshal(podYaml, &podCfg)
	if err != nil {
		log.Println("Failed to Unmarshal the workload.yaml with err %s, exit", err)
		return false
	}

	ns := podCfg.GetNamespace()
	if ns == "" {
		ns = "test"
	}

	_, err = clientSet.CoreV1().Pods(ns).Create(context.TODO(), podCfg, metav1.CreateOptions{})

	if err != nil {
		log.Println("Failed to deploy the workload.yaml pod kind version with err %s, exit\n", err)
		return false
	}

	return true
}
