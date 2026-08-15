variable "region" {
  type        = string
  description = "AWS region"
}
variable "cluster_name" {
  type        = string
  description = "EKS cluster name"
}
variable "kubernetes_version" {
  type    = string
  default = "1.29"
}
variable "vpc_cidr" {
  type    = string
  default = "10.20.0.0/16"
}
variable "tags" {
  type    = map(string)
  default = {}
}
