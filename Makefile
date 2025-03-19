## test: runs all tests
test:
		@go test -v ./...
		
## cover: opens coverage in browser
cover:
		@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

## Coverage: displays test coverage
coverage:
		@go test -cover ./...

## build_cli: builds the command line tool devify and copies it to myapp
build_cli:
		@go build -o ../myapp/devify ./cmd/cli






##################################
# Dev-Only Targets
##################################
.PHONY: stage-all
stage-all:
	@echo "Staging all files..."
	@git add .
	@echo "All files staged!"

.PHONY: unstage-all
unstage-all:
	@echo "Unstaging all files..."
	@git restore --staged .
	@echo "All files unstaged!"

.PHONY: diff-to-clipboard
diff-to-clipboard:
	@echo "Copying diff to clipboard..."
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

.PHONY: diff
diff: stage-all diff-to-clipboard  unstage-all
	@echo "DIff content on clipboard and ready to paste."

.PHONY: diff-file
diff-file: stage-all
	@echo "Saving diff to diff_output.txt..."
	@git diff --staged > diff_output.txt
	@echo "Diff saved to diff_output.txt."
	@$(MAKE) unstage-all
