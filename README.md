# k8sCRD
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
