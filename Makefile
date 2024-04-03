APP=casdoor
VERSION=1.0.0
GOARCH=amd64
GOPATH=$(shell go env GOPATH)
GOSRC=$(GOPATH)/src
GOFLAGS=-mod=vendor
PREFIX=/usr/local

ifeq ($(TARGET_OS),linux)
    GOOS=linux
    OUTPUT_FOLDER=bin/linux
else ifeq ($(TARGET_OS),darwin)
    GOOS=darwin
    OUTPUT_FOLDER=bin/macos
endif

.PHONY: build clean install uninstall

build:
ifndef TARGET_OS
	$(error TARGET_OS is not set, please specify TARGET_OS=(linux|darwin))
endif
	@mkdir -p $(OUTPUT_FOLDER)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS) -o $(OUTPUT_FOLDER)/$(APP) -ldflags "-X main.version=$(VERSION)" .

clean:
	@rm -rf bin

install: build
	@cp $(OUTPUT_FOLDER)/$(APP) $(PREFIX)/bin/$(APP)
	@echo "Export the binary path by adding the following line to your .bashrc, .bash_profile, or .zshrc:"
	@echo "export PATH=\"$(PREFIX)/bin:\$$PATH\""
	@echo ""
	@echo "For zsh users:"
	@echo "echo 'export PATH=\"$(PREFIX)/bin:\$$PATH\"' >> ~/.zshrc"
	@echo "source ~/.zshrc"
	@echo ""
	@echo "For bash users:"
	@echo "echo 'export PATH=\"$(PREFIX)/bin:\$$PATH\"' >> ~/.bashrc"
	@echo "source ~/.bashrc"

uninstall:
	@rm -f $(PREFIX)/bin/$(APP)
	@echo "To remove the PATH entry, manually delete the following line in your .bashrc, .bash_profile, or .zshrc:"
	@echo "export PATH=\"$(PREFIX)/bin:\$$PATH\""
	@echo ""
	@echo "For zsh users:"
	@echo "Open ~/.zshrc with a text editor, like 'nano ~/.zshrc', and remove the PATH line."
	@echo "After editing, reload the configuration with 'source ~/.zshrc'."
	@echo ""
	@echo "For bash users:"
	@echo "Open ~/.bashrc (or ~/.bash_profile) with a text editor, like 'nano ~/.bashrc', and remove the PATH line."
	@echo "After editing, reload the configuration with 'source ~/.bashrc'."
