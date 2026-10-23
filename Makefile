GO ?= go

.PHONY: test vet generate-check file-cap pre-push

test:
	cd collector && GOWORK=off $(GO) test ./... -count=1
	cd factschema && GOWORK=off $(GO) test ./... -count=1

vet:
	cd collector && GOWORK=off $(GO) vet ./...
	cd factschema && GOWORK=off $(GO) vet ./...

generate-check:
	cd factschema && GOWORK=off $(GO) generate ./...
	git diff --exit-code -- factschema/schema

file-cap:
	@bash scripts/dev/check-file-cap.sh

pre-push: test vet generate-check file-cap
	git diff --check
