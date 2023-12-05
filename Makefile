REPO?=localhost:32000
SKAFFOLD?=skaffold
KUBECTL?=kubectl


dev:
	$(SKAFFOLD) run -p https
	SKAFFOLD_DEFAULT_REPO=$(REPO) $(SKAFFOLD) run --port-forward --tail


info:
	@echo create a client and then swap its credentials in the iam configmap 
