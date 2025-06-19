PHONY: help
.DEFAULT_GOAL := help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

git_tag_increment_minor_version: ## Release with increment minor tag version
	@TAG_ON_HEAD_COMMIT=$$(git tag --points-at HEAD); \
	if [ -n "$$TAG_ON_HEAD_COMMIT" ]; \
	then \
			echo "Tag '$$TAG_ON_HEAD_COMMIT' already exists on that commit. Skipping tagging."; \
	else \
			echo "Getting last tag..."; \
			LAST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
			echo "Last tag: $$LAST_TAG"; \
			MAJOR=$$(echo $$LAST_TAG | cut -d. -f1); \
			MINOR=$$(echo $$LAST_TAG | cut -d. -f2); \
			PATCH=$$(echo $$LAST_TAG | cut -d. -f3); \
			NEW_PATCH=$$((PATCH + 1)); NEW_TAG="$${MAJOR}.$${MINOR}.$${NEW_PATCH}"; \
			echo "New tag: $$NEW_TAG"; \
			git tag "$$NEW_TAG"; \
			git push origin "$$NEW_TAG"; \
			echo "Tag '$$NEW_TAG' successfully created and pushed."; \
	fi

run_tests: ## Run tests
	cd tests && go test -v ./...

run_coverage_test: ## Run coverage tests
	cd tests && \
	go test ./... \
      -coverpkg=github.com/KoNekoD/swaglay/pkg/... \
      -coverprofile=coverage.out
	go tool cover -html=tests/coverage.out
