package discovery

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math/rand"
	"strings"
	"time"
)

const (
	defaultK8sDiscoveryNodePortName = "easyraft"
	defaultK8sDiscoverySvcType      = "easyraft"
)

type KubernetesDiscovery struct {
	namespace             string
	matchingServiceLabels map[string]string
	nodePortName          string
	discoveryChan         chan string
	stopChan              chan bool
	delayTime             time.Duration
}

func NewKubernetesDiscovery(namespace string, serviceLabels map[string]string, raftPortName string) DiscoveryMethod {
	if raftPortName == "" {
		raftPortName = defaultK8sDiscoveryNodePortName
	}
	if serviceLabels == nil || len(serviceLabels) == 0 {
		serviceLabels = make(map[string]string)
		serviceLabels["svcType"] = defaultK8sDiscoverySvcType
	}
	delayTime := time.Duration(rand.Intn(30)+5) * time.Second
	return &KubernetesDiscovery{
		namespace:             namespace,
		matchingServiceLabels: serviceLabels,
		nodePortName:          raftPortName,
		discoveryChan:         make(chan string),
		stopChan:              make(chan bool),
		delayTime:             delayTime,
	}
}

func (k *KubernetesDiscovery) Start(_ string, _ int) (chan string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	go k.discovery(clientSet)
	return k.discoveryChan, nil
}

func (k *KubernetesDiscovery) discovery(clientSet *kubernetes.Clientset) {
	for {
		select {
		case <-k.stopChan:
			return
		default:
			services, err := clientSet.CoreV1().Services(k.namespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(k.matchingServiceLabels).String(),
				Watch:         false,
			})
			if err != nil {
				log.Println(err)
				continue
			}

			for _, svc := range services.Items {
				set := labels.Set(svc.Spec.Selector)
				listOptions := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(set).String(),
				}
				pods, err := clientSet.CoreV1().Pods(svc.Namespace).List(context.Background(), listOptions)
				if err != nil {
					log.Println(err)
					continue
				}
				for _, pod := range pods.Items {
					if strings.ToLower(string(pod.Status.Phase)) == "running" {
						podIp := pod.Status.PodIP
						var raftPort v1.ContainerPort
						for _, container := range pod.Spec.Containers {
							for _, port := range container.Ports {
								if port.Name == k.nodePortName {
									raftPort = port
									break
								}
							}
						}
						if podIp != "" && raftPort.ContainerPort != 0 {
							k.discoveryChan <- fmt.Sprintf("%v:%v", podIp, raftPort.ContainerPort)
						}

					}
				}
			}

			time.Sleep(k.delayTime)
		}
	}
}

func (k *KubernetesDiscovery) SupportsNodeAutoRemoval() bool {
	return true
}

func (k *KubernetesDiscovery) Stop() {
	k.stopChan <- true
	close(k.discoveryChan)
}
