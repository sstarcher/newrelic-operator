
IMAGE = sstarcher/newrelic-operator:test

.PHONY: all build

all: e2e

dep:
	dep ensure

build: dep
ifdef NEW_RELIC_APIKEY
	git checkout deploy/operator.yaml
	git checkout deploy/role_binding.yaml
	@sed -i '' s/REPLACE_ME_NEW_RELIC_APIKEY/${NEW_RELIC_APIKEY}/ deploy/operator.yaml
	@sed -i '' s/REPLACE_NAMESPACE/default/ deploy/role_binding.yaml
else
	echo "Missing NEW_RELIC_APIKEY variable"
	exit 1
endif
	operator-sdk build $(IMAGE)

push: build
	docker push $(IMAGE)

test: dep
	go test `go list ./... | grep -v vendor | grep -v e2e | grep -v newrelic-operator/newrelic-operator`
	
e2e: build
	operator-sdk test local --kubeconfig=${HOME}/.kube/config ./test/e2e --go-test-flags='-v' --namespace=default


