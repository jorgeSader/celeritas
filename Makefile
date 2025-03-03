## test: runs all tests
test:
		@go test -v ./...
		
## cover: opens coverage in browser
cover:
		@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

## Coverage: displays test coverage
coverage:
		@go test -cover ./...

## build_cli: builds the command line tool celeritas and copies it to myapp
build_cli:
		@go build -o ../myapp/celeritas ./cmd/cli






##################################
# For Dev Only
##################################
.PHONY: stage-all
stage-all:
	@echo "Staging all files..."
	git add .
	@echo "All file Staged"

.PHONY: diff
diff:
	@echo "Creating diff file..."
	git diff --staged > changes.diff
	@echo "Diff file Created"

.PHONY: diff-all
diff-all: stage-all diff
	@echo "Staged all modified files and created a 'changes.diff' file."