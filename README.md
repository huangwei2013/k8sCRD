# k8sCRD
kubernetes的CRD实现示例

## crdAPIDemo
基于 kubebuilder 生成的默认 operator 工程
* 扩展了 Deployment 和 Service 功能

### 使用
下载并放置在有效golang路径下
```
go mod tidy
make
make install & make run
```