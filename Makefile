# Copyright (c) 2020 Learning by Example maintainers.
#
#  Permission is hereby granted, free of charge, to any person obtaining a copy
#  of this software and associated documentation files (the "Software"), to deal
#  in the Software without restriction, including without limitation the rights
#  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#  copies of the Software, and to permit persons to whom the Software is
#  furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included in
#  all copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
#  THE SOFTWARE.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
COVERAGE=$(GOTOOL) cover
GOFORMAT=$(GOCMD) fmt
GORUN=$(GOCMD) run
GOGET=$(GOCMD) get
BUILD_DIR=build
SCRIPTS_DIR=scripts
BINARY_NAME=$(BUILD_DIR)/go-microservice
APP_PATH="./internal/app"
default: build

build: clean cpycfg test
	$(GOBUILD) -o $(BINARY_NAME) -v $(APP_PATH)
build-no-test: clean cpycfg
	$(GOBUILD) -o $(BINARY_NAME) -v $(APP_PATH)
test:
	$(GOTEST) -short -v -cover -coverprofile=coverage.out -covermode=atomic $(APP_PATH)/...
integration:
	$(GOTEST) -v -cover -coverprofile=coverage.out -covermode=atomic $(APP_PATH)/...
coverage: test
	$(COVERAGE) -html=coverage.out
clean:
	$(GOCLEAN) $(APP_PATH)
	rm -rf $(BUILD_DIR)
format:
	$(GOFORMAT) $(APP_PATH)/...
cpycfg:
	mkdir $(BUILD_DIR)
	mkdir $(BUILD_DIR)/config
	cp config/*.* $(BUILD_DIR)/config/
run: build
	./$(BINARY_NAME)
run-postgresql: build
	./$(BINARY_NAME) -config $(BUILD_DIR)/config/postgresql.json
docker:
	./$(SCRIPTS_DIR)/docker.sh
deploy: docker
	./$(SCRIPTS_DIR)/deploy.sh

update:
	$(GOGET) -u all
