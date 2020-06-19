package main

import (
	"flag"
	"fmt"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	//	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// K8S Informer records
type K8sRecord struct {
	gorm.Model
	EventType string `gorm:"index:event_type"`
	ActType string
	EventID string `gorm:"index:event_id"`
	EventName string `gorm:"index:event_name"`
	PodIP     string `gorm:"index:pod_id"`
	PodName   string
	Cluster   string
	Namespace string
	Action    string
	Kind      string
	Type      string
	Source    string
	Name      string
	Reason    string
	Message   string
	Status    string
	Phase     string
	FirstTimestamp time.Time
	LastTimestamp time.Time
}


func main() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open("mysql", "root:123456@(39.100.133.126:32307)/stbtest?charset=utf8&parseTime=true&loc=Local")
	if err != nil {
		fmt.Println("connect db error: ", err)
	}
	defer db.Close()

	if db.HasTable(&K8sRecord{}) { //判断表是否存在
		db.AutoMigrate(&K8sRecord{}) //存在就自动适配表，也就说原先没字段的就增加字段
	} else {
		db.CreateTable(&K8sRecord{}) //不存在就创建新表
	}

	// 初始化 client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	// 初始化 informer
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	defer runtime.HandleCrash()

	// 启动 informer，list & watch
	go informerFactory.Start(stopCh)
	informerHandler(informerFactory, stopCh, db)

	<-stopCh
}

func informerHandler(informerFactory informers.SharedInformerFactory, stopCh <-chan struct{}, db *gorm.DB){

	informersMap := make(map[string]interface{})
	informersMap["node"] = informerFactory.Core().V1().Nodes()
	informersMap["event"] = informerFactory.Core().V1().Events()
	informersMap["namespace"] = informerFactory.Core().V1().Namespaces()
	informersMap["pod"] = informerFactory.Core().V1().Pods()

	defaultHandler := func(name string, db *gorm.DB) cache.ResourceEventHandler {
		return cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {  parseInform4K8SRecord(name, "add", obj, db); },
			UpdateFunc: func(old, new interface{}) {
				parseInform4K8SRecord(name, "update.old", old, db);
				parseInform4K8SRecord(name, "update.new", new, db);
			},
			DeleteFunc: func(obj interface{}) {parseInform4K8SRecord(name, "delete", obj, db); },
		}
	}

	for name, informer := range informersMap {
		fmt.Println(" informer handlers : ",name)

		switch infromer := informer.(type){
			case  coreinformers.NodeInformer:
				infromer.Informer().AddEventHandler(defaultHandler(name,db))
			case  coreinformers.EventInformer:
				infromer.Informer().AddEventHandler(defaultHandler(name,db))
			case  coreinformers.NamespaceInformer:
				infromer.Informer().AddEventHandler(defaultHandler(name,db))
			case  coreinformers.PodInformer:
				infromer.Informer().AddEventHandler(defaultHandler(name,db))
			default :continue;
		}

	}

	informerFactory.Start(stopCh)
	//informerFactory.WaitForCacheSync(stopCh)

}

// TODO：这部分要做大量字段枚举
func parseInform4K8SRecord(informType string, actType string, obj interface{}, db *gorm.DB){
	fmt.Println(obj);

	var k8sRecord K8sRecord
	switch informType {
	case "event":
		node := obj.(*corev1.Event)
		k8sRecord = K8sRecord{EventType: informType, ActType: actType,
			FirstTimestamp : node.FirstTimestamp.Time, LastTimestamp : node.LastTimestamp.Time,
			Type : node.Type, EventID: fmt.Sprint(node.UID), Cluster: node.ClusterName, Kind: node.Kind,
			Name: node.Name,Namespace: node.Namespace, Action: node.Action, Phase: "",
			Reason: node.Reason, Source : node.Source.String(),
			Message:node.Message}
	case "namespace":
		node := obj.(*corev1.Namespace)
		k8sRecord = K8sRecord{EventType: informType, ActType: actType,
			FirstTimestamp: node.CreationTimestamp.Time,
			EventID: fmt.Sprint(node.UID), Cluster: node.ClusterName, Kind: node.Kind,
			Name: node.Name, Namespace: node.Name, Action: "", Phase: fmt.Sprint(node.Status.Phase),
			Reason: "", Source : "",
			Message:""}
	case "node":
		node := obj.(*corev1.Node)
		k8sRecord = K8sRecord{EventType: informType, ActType: actType,
			FirstTimestamp: node.CreationTimestamp.Time,
			EventID: fmt.Sprint(node.UID), Cluster: node.ClusterName, Kind: node.Kind,
			Name: node.Name, Namespace: node.Namespace, Action: "", Phase: fmt.Sprint(node.Status.Phase),
			Reason: "", Source : "",
			Message:""}
	case "pod":
		node := obj.(*corev1.Pod)
		k8sRecord = K8sRecord{EventType: informType, ActType: actType,
			FirstTimestamp: node.CreationTimestamp.Time,
			EventID: fmt.Sprint(node.UID), Cluster: node.ClusterName, Kind: node.Kind,
			PodName: node.Name, Namespace: node.Namespace, Action: "", Phase: fmt.Sprint(node.Status.Phase),
			Reason: "", Source : "",
			Message:""}
	default:
		return
	}

	db.Create(&k8sRecord)
}
