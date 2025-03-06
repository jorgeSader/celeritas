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
# Dev-Only Targets
##################################
.PHONY: stage-all
stage-all:
	@echo "Staging all files..."
	git add .
	@echo "All files staged!"

.PHONY: diff
diff:
	@echo "Copying diff to clipboard..."
	@# Detect OS and use appropriate clipboard tool
	@if [ "$$(uname)" = "Darwin" ]; then \
		git diff --staged | pbcopy; \
		echo "Diff copied to clipboard (macOS)"; \
	elif [ "$$(uname)" = "Linux" ]; then \
		if command -v xclip >/dev/null 2>&1; then \
			git diff --staged | xclip -selection clipboard; \
			echo "Diff copied to clipboard (Linux/xclip)"; \
		elif command -v wl-copy >/dev/null 2>&1; then \
			git diff --staged | wl-copy; \
			echo "Diff copied to clipboard (Linux/wl-copy)"; \
		else \
			echo "Error: Install xclip or wl-copy for clipboard support"; \
			exit 1; \
		fi; \
	elif [ "$$(uname -o 2>/dev/null)" = "Msys" ] || [ "$$(uname -o 2>/dev/null)" = "Cygwin" ]; then \
		git diff --staged | clip; \
		echo "Diff copied to clipboard (Windows)"; \
	else \
		echo "Error: Unsupported OS for clipboard copy"; \
		exit 1; \
	fi

.PHONY: diff-all
diff-all: stage-all diff
	@echo "Staged all modified files and copied diff to clipboard."