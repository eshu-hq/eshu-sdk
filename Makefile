GO ?= go

# Coverage floors enforced by the race lanes in CI (see coverage-floors in
# .github/workflows/ci.yml) and by `make race` below. Floors start at the
# measured baselines and only rise; neither module closes below 90% per
# epic #27. Keep the Makefile defaults in sync with the workflow inputs.
COVERAGE_FLOOR_COLLECTOR ?= 85.0
COVERAGE_FLOOR_FACTSCHEMA ?= 48.3
COVERAGE_FLOOR_WIRECONTRACT ?= 84.7

.PHONY: build test vet lint tidy-check generate-check file-cap race coverage govulncheck shellcheck shape actionlint gosec ci pre-push

build:
	cd collector && GOWORK=off $(GO) build ./...
	cd factschema && GOWORK=off $(GO) build ./...
	cd wirecontract && GOWORK=off $(GO) build ./...

test:
	cd collector && GOWORK=off $(GO) test ./... -count=1
	cd factschema && GOWORK=off $(GO) test ./... -count=1
	cd wirecontract && GOWORK=off $(GO) test ./... -count=1

vet:
	cd collector && GOWORK=off $(GO) vet ./...
	cd factschema && GOWORK=off $(GO) vet ./...
	cd wirecontract && GOWORK=off $(GO) vet ./...

lint:
	cd collector && GOWORK=off golangci-lint run ./...
	cd factschema && GOWORK=off golangci-lint run ./...
	cd wirecontract && GOWORK=off golangci-lint run ./...

tidy-check:
	bash scripts/ci/tidy-check.sh collector
	bash scripts/ci/tidy-check.sh factschema
	bash scripts/ci/tidy-check.sh wirecontract

generate-check:
	cd factschema && GOWORK=off $(GO) generate ./...
	@status="$$(git status --porcelain --untracked-files=all --ignored=matching -- factschema/schema)" || exit $$?; \
	if [ -n "$$status" ]; then \
		printf '%s\n' 'Generated schemas differ from committed artifacts:' "$$status" >&2; exit 1; \
	fi

file-cap:
	@bash scripts/dev/check-file-cap.sh

race:
	cd collector && GOWORK=off $(GO) test -race ./... -count=1 -coverprofile=coverage.out
	cd collector && bash ../scripts/ci/check-coverage.sh coverage.out $(COVERAGE_FLOOR_COLLECTOR)
	cd factschema && GOWORK=off $(GO) test -race ./... -count=1 -coverprofile=coverage.out
	cd factschema && bash ../scripts/ci/check-coverage.sh coverage.out $(COVERAGE_FLOOR_FACTSCHEMA)
	cd wirecontract && GOWORK=off $(GO) test -race ./... -count=1 -coverprofile=coverage.out
	cd wirecontract && bash ../scripts/ci/check-coverage.sh coverage.out $(COVERAGE_FLOOR_WIRECONTRACT)

coverage: race

govulncheck:
	cd collector && GOWORK=off govulncheck ./...
	cd factschema && GOWORK=off govulncheck ./...
	cd wirecontract && GOWORK=off govulncheck ./...

shellcheck:
	git ls-files '*.sh' | xargs shellcheck --severity=warning

# Shape tests for the coverage gate helper and the wrapper contract: they
# run in `make ci` and in the CI `shape` job, so a gutted helper cannot go
# green by being skipped (issue #34). The lane-selection shape test is
# retired with #39: the classifier it pinned now lives in the org reusable,
# and this workflow carries no classifier left to test.
shape:
	bash scripts/ci/test-coverage-floor.sh
	bash scripts/ci/test-floors-sync.sh
	bash scripts/ci/test-wrapper-inputs.sh

actionlint:
	actionlint -no-color .github/workflows/*.yml

# Advisory until the backlog is triaged and measured: records SARIF, never
# fails the local run, mirroring the CI gosec lane.
gosec:
	cd collector && GOWORK=off gosec -fmt sarif -out ../gosec-collector.sarif ./... || true
	cd factschema && GOWORK=off gosec -fmt sarif -out ../gosec-factschema.sarif ./... || true
	cd wirecontract && GOWORK=off gosec -fmt sarif -out ../gosec-wirecontract.sarif ./... || true
	@printf '%s\n' 'gosec: advisory only, see gosec-*.sarif (untracked, do not commit)'

# Full local parity with the CI lanes (minus the satellite consumer-compat
# run, which needs network clones; run scripts/ci/consumer-compat.sh for it).
ci: build test vet lint tidy-check generate-check file-cap race govulncheck shellcheck shape actionlint gosec
	git diff --check

pre-push: build test vet lint tidy-check generate-check file-cap
	git diff --check
