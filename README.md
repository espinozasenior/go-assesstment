### PRE-REQUISITES
* [Golang version v1.23.0+](https://go.dev/doc/install)
* [Gcloud CLI](https://cloud.google.com/sdk/docs/install)
* [Terraform](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/install-cli)
* [Docker version 17.03+]()
* [Kubectl version v1.11.3+](https://kubernetes.io/docs/tasks/tools/)
* [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

### INSTRUCTIONS TO CLUSTER DEPLOYMENT
- Copy `terraform/terraform.tfvars.example` to `terraform/terraform.tfvars` and fill in your GCP project ID
- Run `make init` to initialize Terraform
- Run `make plan` to see any changes that are required for your infrastructure
- Run `make apply` to create the GKE cluster
- Run `make kubeconfig` to generate the kubeconfig file to interact with your cluster

### INSTRUCTIONS TO APPLICATION DEPLOYMENT
- Run `make install` to install to apply the CRD definition, this will register our custom resource type with the Kubernetes API server.
- Run `make run` to build the operator binary and run it locally.
- Run `make deploy` to deploy the sample AppDeployment resource to the cluster.
- Run `kubectl get appdeployments.platform.deskree.com` to see the status of the AppDeployment resource.

### TROUBLESHOOTING
"google: could not find default credentials" - run `gcloud auth application-default login`

"oauth2: invalid_grant" - run `gcloud auth application-default login`

"gke-gcloud-auth-plugin was not found or is not executable" - run `gcloud components install gke-gcloud-auth-plugin`