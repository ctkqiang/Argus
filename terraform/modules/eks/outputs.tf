output "cluster_endpoint" {
  value       = "TBD: EKS cluster endpoint"
  description = "EKS control plane endpoint"
}
output "cluster_ca" {
  value       = "TBD: EKS CA data"
  description = "Base64 encoded cluster CA cert"
}
output "oidc_provider_arn" {
  value       = "TBD: OIDC"
  description = "OIDC provider ARN for IRSA"
}
