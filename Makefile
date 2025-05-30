MODULES=$(shell find . -mindepth 2 -maxdepth 4 -type f -name 'go.mod' | cut -c 3- | sed 's|/[^/]*$$||' | sort -u | tr / :)

build:
	$(MAKE) $(addprefix build-, $(MODULES))

build-%:
	cd $(subst :,/,$*); go build .

tidy:
	$(MAKE) $(addprefix tidy-, $(MODULES))

tidy-%:
	cd $(subst :,/,$*); go mod tidy

test:
	$(MAKE) $(addprefix test-, $(MODULES))

test-%:
	cd $(subst :,/,$*); go test ./... -coverprofile cover.out ;

lint:
	$(MAKE) $(addprefix lint-, $(MODULES))

lint-fix:
	$(MAKE) $(addprefix lint-fix-, $(MODULES))

lint-%:
	cd $(subst :,/,$*); golangci-lint run ./...

lint-fix-%:
	cd $(subst :,/,$*); golangci-lint run --fix ./...

scan:
	$(MAKE) $(addprefix scan-, $(MODULES))

scan-%:
	cd $(subst :,/,$*); syft scan . -o json | grype --fail-on low

check-go-version-consistency:
	hack/check_go_version.sh

readme:
	cd internal/readme; go run main.go
