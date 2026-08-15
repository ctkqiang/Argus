# templates/crds/

CRD YAML 由 `make crd`（controller-gen）生成后放这里。**禁止手动改**。

Helm 安装 CRD 有两条路：
1. 默认：`Values.crds.install=true`，Helm 走 templates/crds 安装。
2. 想自己管 CRD 生命周期（比如升级前先看 diff）：把 Values.crds.install=false 改成 false，
   然后 `kubectl apply -f deploy/helm/argus/templates/crds/`。
