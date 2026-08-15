terraform {
  required_version = ">= 1.5.0"
  required_providers {
    helm = { source = "hashicorp/helm", version = ">= 2.11" }
  }
}

resource "helm_release" "argus" {
  name             = var.release_name
  namespace        = var.namespace
  create_namespace = true
  chart            = "${path.module}/../../../deploy/helm/argus"
  values           = [yamlencode(var.values)]
  atomic           = true
  timeout          = 600
}
