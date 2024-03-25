MODULES=$(shell find . -mindepth 2 -maxdepth 4 -type f -name 'go.mod' | cut -c 3- | sed 's|/[^/]*$$||' | sort -u | tr / :)

build:
	$(MAKE) $(addprefix build-, $(MODULES))

build-%:
	cd $(subst :,/,$*); go build .

tidy:
	$(MAKE) $(addprefix tidy-, $(MODULES))

tidy-%:
	cd $(subst :,/,$*); go mod tidy

fmt:
	$(MAKE) $(addprefix fmt-, $(MODULES))

fmt-%:
	cd $(subst :,/,$*); go fmt ./...

vet:
	$(MAKE) $(addprefix vet-, $(MODULES))

vet-%:
	cd $(subst :,/,$*); go vet ./... ;\

test:
	$(MAKE) $(addprefix test-, $(MODULES))

test-%:
	cd $(subst :,/,$*); go test ./... -coverprofile cover.out ;

lint:
	revive -config revive.toml ./...

scan:
	$(MAKE) $(addprefix scan-, $(MODULES))

scan-%:
	cd $(subst :,/,$*); syft scan . -o json | grype --fail-on low
