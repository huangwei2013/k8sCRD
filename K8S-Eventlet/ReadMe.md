
## 事件类型

NewSharedInformerFactory.Core().V1()包含的informer类型：

```
	type Interface interface {
		ComponentStatuses() ComponentStatusInformer
		ConfigMaps() ConfigMapInformer
		Endpoints() EndpointsInformer
		Events() EventInformer
		LimitRanges() LimitRangeInformer
		Namespaces() NamespaceInformer
		Nodes() NodeInformer
		PersistentVolumes() PersistentVolumeInformer
		PersistentVolumeClaims() PersistentVolumeClaimInformer
		Pods() PodInformer
		PodTemplates() PodTemplateInformer
		ReplicationControllers() ReplicationControllerInformer
		ResourceQuotas() ResourceQuotaInformer
		Secrets() SecretInformer
		Services() ServiceInformer
		ServiceAccounts() ServiceAccountInformer
	}
```