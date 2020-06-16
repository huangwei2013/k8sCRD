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
	EventName string `gorm:"index:event_name"`
	PodIP     string `gorm:"index:pod_id"`
	PodName   string
	Cluster   string
	Namespace string
	Action    string
	Kind      string
	Source    string
	Name      string
	Reason    string
	Message   string
	Status    string
	Phase     string
	FirstTimestamp time.Time
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
	syncPeriod := time.Duration(100 * 1000000000)
	fmt.Println(syncPeriod)
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	defer runtime.HandleCrash()

	// 启动 informer，list & watch
	go informerFactory.Start(stopCh)

	informerHanlder(informerFactory, stopCh, db)

	<-stopCh
}

func informerHanlder(informerFactory informers.SharedInformerFactory, stopCh <-chan struct{}, db *gorm.DB){
	nodeInformer := informerFactory.Core().V1().Nodes()
	eventsInformer := informerFactory.Core().V1().Events()
	nameInformer := informerFactory.Core().V1().Namespaces()
	podInformer := informerFactory.Core().V1().Pods()

	informerNamespace(nameInformer, db)
	informerEvent(eventsInformer, db)
	informerPod(podInformer)
	informerNode(nodeInformer)

	go nodeInformer.Informer().Run(stopCh)
	go eventsInformer.Informer().Run(stopCh)
	go nameInformer.Informer().Run(stopCh)
	go podInformer.Informer().Run(stopCh)
}

// TODO：这部分要做大量字段枚举
func parseInform4K8SRecord(informType string, obj interface{}, db *gorm.DB){

	var k8sRecord K8sRecord
	switch informType {
	case "event":
		node := obj.(*corev1.Event)
		k8sRecord = K8sRecord{EventType: informType,
			FirstTimestamp : node.FirstTimestamp.Time,
			Name: node.Name,Namespace: node.Namespace, Action: node.Action, Phase: "",
			Reason: node.Reason, Kind: node.Kind, Source : node.Source.String(),
			Message:node.Message, Cluster:node.ClusterName}
	case "namespace":
		node := obj.(*corev1.Namespace)
		k8sRecord = K8sRecord{EventType: informType,
			FirstTimestamp: node.ObjectMeta.CreationTimestamp.Time,
			Name: node.Name, Namespace: node.Name, Action: "", Phase: fmt.Sprint(node.Status.Phase),
			Reason: "", Kind: node.Kind, Source : "",
			Message:"", Cluster:node.ClusterName}
	default:
		return
	}

	db.Create(&k8sRecord)
}

func informerEvent(informer coreinformers.EventInformer, db *gorm.DB){
	fmt.Println("Event informer handlers")

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}){ fmt.Println(obj);   parseInform4K8SRecord("event", obj, db); },
		UpdateFunc: func(old, new interface{}){ fmt.Println(old); fmt.Println(new); parseInform4K8SRecord("event", new, db);},
		DeleteFunc: func(obj interface{}){ fmt.Println(obj); parseInform4K8SRecord("event", obj, db);},
	})
}

func informerNamespace(informer coreinformers.NamespaceInformer, db *gorm.DB){
	fmt.Println("Namespace informer handlers")

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}){ fmt.Println(obj);   parseInform4K8SRecord("namespace", obj, db); },
		UpdateFunc: func(old, new interface{}){ fmt.Println(old); fmt.Println(new); parseInform4K8SRecord("namespace", new, db);},
		DeleteFunc: func(obj interface{}){ fmt.Println(obj); parseInform4K8SRecord("namespace", obj, db);},
	})
}

func informerPod(informer coreinformers.PodInformer){
	fmt.Println("Pod informer handlers")

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}){ fmt.Println(obj)},
		UpdateFunc: func(old, new interface{}){ fmt.Println(old); fmt.Println(new)},
		DeleteFunc: func(obj interface{}){ fmt.Println(obj)},
	})
}

func informerNode(informer coreinformers.NodeInformer){
	fmt.Println("Node informer handlers")

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAddNode,
		UpdateFunc: onUpdateNode,
		DeleteFunc: onDeleteNode,
	})

}

func onAddNode(obj interface{}) {
	node := obj.(*corev1.Node)
	fmt.Println("add a node:", node.Name)
}

func onUpdateNode(old, new interface{} ) {
	// 此处省略 workqueue 的使用
	oldNode := old.(*corev1.Node)
	newNode := new.(*corev1.Node)
	fmt.Println("update a node:", oldNode.Name, newNode.Name)
}

func onDeleteNode(obj interface{}) {
	node := obj.(*corev1.Node)
	fmt.Println("delete a node:", node.Name)
}