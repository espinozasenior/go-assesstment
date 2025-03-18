### REQUIRED DEPENDENCIES 
* [Golang](https://go.dev/doc/install)
* [Gcloud CLI](https://cloud.google.com/sdk/docs/install)
* [Terraform](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/install-cli)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Kubebuilder](https://book.kubebuilder.io/quick-start)

## INSTRUCTIONS

### GKE CLUSTER SETUP

- Navigate to terraform directory `cd /path/to/go-assesstment/terraform`
- Copy `terraform/terraform.tfvars.example` to `terraform/terraform.tfvars` and fill in your GCP project ID
- Run `make init` to initialize Terraform
- Run `make plan` to see any changes that are required for your infrastructure
- Run `make apply` to create the GKE cluster
- Run `make kubeconfig` to generate the kubeconfig file to interact with your cluster

### RUN THE APPLICATION 

### CLI USAGE
#### Building the CLI
```
# Navigate to the project root directory
cd /path/to/go-assesstment

# Build the CLI tool
go build -o go-assessment ./cmd/cli
```

#### CLI Commands

**Login**
```
./go-assessment login --username <your-username> --password <your-password>
```

**Deploy an Application**
```
./go-assessment deploy --name <app-name> --image <container-image> --memoryLimit <memory-limit> --minReplicas <min-replicas> --maxReplicas <max-replicas>
```
Example:
```
./go-assessment deploy --name myapp --image nginx:latest --memoryLimit 512Mi --minReplicas 1 --maxReplicas 3
```

**Check Deployment Status**
```
./go-assessment status --name <app-name>
```

**Destroy a Deployment**
```
./go-assessment destroy --name <app-name>
```

### TROUBLESHOOTING
"google: could not find default credentials" - run `gcloud auth application-default login`

"oauth2: invalid_grant" - run `gcloud auth application-default login`

"gke-gcloud-auth-plugin was not found or is not executable" - run `gcloud components install gke-gcloud-auth-plugin`