package workload

import (
	"context"
	"log"
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Config represents the structure of the YAML data
type Config struct {
	Data map[string]string `yaml:"data"`
}

// the function will read a configmap file that contains workload.yaml
// Then generate the workload.yaml in the directory
func PorcessWorkloadCfgfile(cmPath string, workloadPath string) bool {
	cmFile, err := os.ReadFile(cmPath)
	if err != nil {
		log.Printf("Read configmap file %s does not exist", cmPath)
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
	err = os.WriteFile(workloadPath, []byte(workload), 0643)
	if err != nil {
		log.Printf(" write workload.yaml failed")
		return false
	}

	return true
}

// the function will delete the existing workload pods under test ns
func DeleteWorkloadPod(ctx context.Context, clientSet *dynamic.DynamicClient, wait bool) bool {

	podName := "workload"
	ns := "test"

	// Define the GVR
	podGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	err := clientSet.Resource(podGVR).Namespace(ns).Delete(context.TODO(), podName, metav1.DeleteOptions{})

	if err != nil && err.Error() == "pods \"workload\" not found" {
		return true
	} else if err != nil {
		log.Printf(" Delete workload pod failed err=%s \n", err)
		return false
	}
	//waiting for the pods deleted
	if wait == false {
		return true
	}

	for i := 0; i < 15; i++ {
		_, err := clientSet.Resource(podGVR).Namespace(ns).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil && err.Error() == "pods \"workload\" not found" {
			return true
		}
		time.Sleep(time.Duration(1+i*10) * time.Second)
	}

	return false
}

func ApplyWorkloadPod(ctx context.Context, clientSet *dynamic.DynamicClient, workloadPath string) bool {
	podYaml, err := os.ReadFile(workloadPath)

	if err != nil {
		log.Println("Failed to read the workload.yaml with err %s, exit", err)
		return false
	}

	var podCfg map[string]interface{}
	err = yaml.Unmarshal(podYaml, &podCfg)
	if err != nil {
		log.Println("Failed to Unmarshal the workload.yaml with err %s, exit", err)
		return false
	}

	pod := &unstructured.Unstructured{Object: podCfg}

	// Define the GVR
	podGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	_, err = clientSet.Resource(podGVR).Namespace("test").Create(context.TODO(), pod, metav1.CreateOptions{})

	if err != nil {
		log.Println("Failed to deploy the workload.yaml pod kind version with err %s, exit\n", err)
		return false
	}

	return true
}
