.PHONY: init plan apply destroy kubeconfig

init:
	cd terraform && terraform init

plan:
	cd terraform && terraform plan -var-file=terraform.tfvars

apply:
	cd terraform && terraform apply -var-file=terraform.tfvars -auto-approve

destroy:
	cd terraform && terraform destroy -var-file=terraform.tfvars -auto-approve

kubeconfig:
	./scripts/generate_kubeconfig.sh