module github.com/defenseunicorns/pkg/exec

go 1.24.0

replace github.com/defenseunicorns/pkg/helpers/v2 => ../helpers

require (
	github.com/defenseunicorns/pkg/helpers/v2 v2.0.1
	golang.org/x/sync v0.13.0
)

require (
	github.com/otiai10/copy v1.14.1 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	golang.org/x/sys v0.32.0 // indirect
	oras.land/oras-go/v2 v2.5.0 // indirect
)
