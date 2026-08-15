terraform {
  required_version = ">= 1.5.0"
  backend "s3" {} # 通过 -backend-config=../../backends/dev.<cloud>.hcl 在 init 时注入
}

# Dev 最小例子：默认走 modules/eks。换云就换成 modules/gke / modules/aks。
# provider "aws" { region = "us-east-1" }
# module "cluster" { source = "../../modules/eks" ... }

# 然后用 argus-helm 把 Chart 装进去。
# provider "helm" { kubernetes { ... } }
# module "argus" {
#   source = "../../modules/argus-helm"
#   values = { runMode = { default = "enforce", defaultFailureMode = "fail-open" } }
# }
