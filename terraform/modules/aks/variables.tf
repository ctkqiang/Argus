variable "resource_group_name" {
  type        = string
  description = "Existing RG name"
}
variable "location" {
  type        = string
  description = "Azure region"
}
variable "cluster_name" {
  type        = string
  description = "AKS cluster name"
}
