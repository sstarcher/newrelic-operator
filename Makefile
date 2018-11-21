
IMAGE = sstarcher/newrelic-operator:test

.PHONY: test

test:
	git checkout deploy/operator.yaml
	@sed -i '' s/REPLACE_ME_NEW_RELIC_APIKEY/${NEW_RELIC_APIKEY}/ deploy/operator.yaml
	operator-sdk build $(IMAGE)
	docker push $(IMAGE)
	operator-sdk test local ./test/e2e --go-test-flags='-v' --namespace=default
