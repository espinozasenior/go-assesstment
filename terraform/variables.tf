variable "project_id" {
  description = "The project ID to host the cluster in"
  type        = string
}

variable "region" {
  description = "The region to host the cluster in"
  type        = string
}

variable "cluster_name" {
  description = "The name of the cluster"
  type        = string
}

variable "node_count" {
  description = "Number of nodes in the node pool"
  type        = number
  default     = 2
}

variable "machine_type" {
  description = "The machine type to use for the node pool"
  type        = string
  default     = "e2-medium"
}

variable "disk_size_gb" {
  description = "Size of the node's disk in GB"
  type        = number
  default     = 50
}