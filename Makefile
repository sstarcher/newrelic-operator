
IMAGE = sstarcher/newrelic-operator:test

.PHONY: test

test:
	operator-sdk build $(IMAGE)
	docker push $(IMAGE)
	operator-sdk test local ./test/e2e --go-test-flags='-v' --namespace=default
