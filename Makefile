GOENV=GOOS=darwin GOARCH=amd64
GOCMD=go
GOBUILD=$(GOCMD) build

BINARY_NAME=bin
NAME=web-tail
MAC_DIR=$(NAME)-Darwin-17.7.0
LINUX_DIR=$(NAME)-linux-amd64
OS_DIR=$(MAC_DIR)

TMPLS=tmpls
LOGLI=log.li

# common 
# default is mac
build:
	cd $(BINARY_NAME) && rm -rf $(OS_DIR) && mkdir $(OS_DIR)
	$(GOENV) $(GOBUILD) -o $(BINARY_NAME)/$(OS_DIR)/$(NAME)
	cp -r $(TMPLS) $(BINARY_NAME)/$(OS_DIR)/ && cp $(LOGLI) $(BINARY_NAME)/$(OS_DIR)
	cd $(BINARY_NAME) && zip -rm $(OS_DIR).zip $(OS_DIR)

# linux
build-linux:
	make build GOENV="GOOS=linux GOARCH=amd64" OS_DIR=$(LINUX_DIR)

# mac
build-mac:
	make build

# all
build-all:
	make build-linux
	make build-mac
