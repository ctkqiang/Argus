variable "release_name" {
  type    = string
  default = "argus"
}
variable "namespace" {
  type    = string
  default = "argus"
}
variable "values" {
  type        = any
  default     = {}
  description = "Helm values override map"
}
