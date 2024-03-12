MODULES=$(shell find . -type f -name 'go.mod' -exec dirname {} \; | cut -c 3-)

build:
	$(MAKE) $(addprefix build-, $(MODULES))

build-%:
	cd $*; go build .

tidy:
	$(MAKE) $(addprefix tidy-, $(MODULES))

tidy-%:
	cd $*; go mod tidy

fmt:
	$(MAKE) $(addprefix fmt-, $(MODULES))

fmt-%:
	cd $*; go fmt ./...

vet:
	$(MAKE) $(addprefix vet-, $(MODULES))

vet-%:
	cd $*; go vet ./... ;\

test:
	$(MAKE) $(addprefix test-, $(MODULES))

test-%:
	cd $*; go test ./... -coverprofile cover.out ;

lint:
	$(MAKE) $(addprefix lint-, $(MODULES))

lint-%:
	cd $*; revive -config ../revive.toml ./...

scan:
	$(MAKE) $(addprefix scan-, $(MODULES))

scan-%:
	cd $*; syft scan . -o json | grype --fail-on low
