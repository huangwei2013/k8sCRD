# K8S CRD扩展
kubernetes的CRD实现示例

## crdAPIDemo
基于 kubebuilder 生成的默认 operator 工程，增加名为 sloop 的 CRD
* 增加了 Deployment 和 Service 功能，当集群中有 sloop 资源的变动（CRUD），都会触发这个函数进行协调
* 部分代码参照 Operator -- 是的，和 kubebuider 并列的那个 Operator 项目

### 使用
下载并放置在有效golang路径下
```
go mod tidy
make
make install & make run
```

# K8S 事件监听
实时监听 k8s event和namespace、pod等的事件，可以作为监控、系统对接的接入

## K8S-Eventlet
基于client-go、gorm实现，实时监听 & 解析入DB 了多种事件。
代码简明，方便扩展

TODO：没有加入队列机制

### 使用
```
go mod tidy
go run K8S-Eventlet.go
```

### FAQ
go mod 自动依赖可能有问题，若

```
# k8s.io/client-go/tools/clientcmd/api/v1
/go/pkg/mod/k8s.io/client-go@v12.0.0+incompatible/tools/clientcmd/api/v1/conversion.go:29:15: scheme.AddConversionFuncs undefined (type *runtime.Scheme has no field or method AddConversionFuncs)
/go/pkg/mod/k8s.io/client-go@v12.0.0+incompatible/tools/clientcmd/api/v1/conversion.go:31:12: s.DefaultConvert undefined (type conversion.Scope has no field or method DefaultConvert)
```

修改一下 go.mod 中的内容：
```
k8s.io\client-go@v11.0.0+incompatible   换成    k8s.io/client-go v0.18.2
或
k8s.io\client-go@v12.0.0+incompatible   换成    k8s.io/client-go v0.18.2
```
```
