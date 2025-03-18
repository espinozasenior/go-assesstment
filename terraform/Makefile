.PHONY: init plan apply destroy kubeconfig

init:
	terraform init

plan:
	terraform plan -var-file=terraform.tfvars -out=plan.tfplan

apply:
	terraform apply plan.tfplan

destroy:
	terraform destroy -auto-approve

kubeconfig:
	./scripts/generate_kubeconfig.sh