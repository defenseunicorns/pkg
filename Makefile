MODULES=$(shell find . -type f -name 'go.mod' -exec dirname {} \; | cut -c 3-)

PKG?=$*

all: tidy fmt vet test lint

build:
	$(MAKE) $(addprefix tidy-, $(MODULES))

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
	$(MAKE) $(addprefix lint-, $(MODULES))

lint-%:
	cd $(subst :,/,$*); revive -config ../revive.toml ./...
