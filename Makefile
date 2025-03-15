.PHONY: init plan apply destroy kubeconfig

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