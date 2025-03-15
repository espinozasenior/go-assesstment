.PHONY: init plan apply destroy kubeconfig

init:
	terraform init

plan:
	terraform plan -var-file=terraform.tfvars

apply:
	terraform apply -var-file=terraform.tfvars -auto-approve

destroy:
	terraform destroy -var-file=terraform.tfvars -auto-approve

kubeconfig:
	./scripts/generate_kubeconfig.sh