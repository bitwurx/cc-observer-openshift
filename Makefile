.PHONY: build
build:
	@docker run \
		--rm \
		-e CGO_ENABLED=0 \
		-v $(PWD):/usr/src/concord-observer-openshift \
		-w /usr/src/concord-observer-openshift \
		golang /bin/sh -c "go get -v -d && go build -a -installsuffix cgo -o main"
	@docker build -t concord/observer-openshift .
	@rm main

.PHONY: test
test:
	@docker run \
		-d \
		-e CONCORD_STATUS_CHANGE_NOTIFIER_HOST=localhost:5555 \
		-e DEPLOY_NAMESPACE=agilis-dev \
		-e IMAGE_NAMESPACE=agilis-cicd \
		-e OPENSHIFT_API_HOST=svs-rtp-auto-osp-cluster-dev.cisco.com \
		-v $(PWD):/go/src/concord-observer-openshift \
		-v $(PWD)/.src:/go/src \
		-v $(PWD)/token:/var/run/secrets/kubernetes.io/serviceaccount/token \
		-w /go/src/concord-observer-openshift \
		--name concord-observer-openshift_test \
		golang /bin/sh -c "go get -v -t -d && go test -v"
	@docker logs -f concord-observer-openshift_test
	@docker rm -f concord-observer-openshift_test

.PHONY: test-short
test-short:
	@docker run \
		--rm \
		-it \
		-e CONCORD_STATUS_CHANGE_NOTIFIER_HOST=localhost:5555 \
		-v $(PWD):/go/src/concord-observer-openshift \
		-v $(PWD)/.src:/go/src \
		-w /go/src/concord-observer-openshift \
		golang /bin/sh -c "go get -v -t -d && go test -short -v -coverprofile=.coverage.out"
