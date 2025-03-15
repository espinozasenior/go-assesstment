.PHONY: init plan apply destroy kubeconfig build deploy install run

init:
	cd terraform && terraform init

plan:
	cd terraform && terraform plan -var-file=terraform.tfvars -out=plan.tfplan

apply:
	cd terraform && terraform apply plan.tfplan

destroy:
	cd terraform && terraform destroy -auto-approve

kubeconfig:
	./scripts/generate_kubeconfig.sh

# Operator commands
build:
	go build -o bin/app-operator main.go

# Install CRD definition
install:
	kubectl apply -f config/crd/platform.deskree.com_appdeployments.yaml

# Run the operator locally
run: build
	./bin/app-operator

# Deploy a sample AppDeployment
deploy: install
	kubectl apply -f config/crd/app-deployment.yaml