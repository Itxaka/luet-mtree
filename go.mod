module github.com/itxaka/luet-mtree

go 1.16

require (
	github.com/docker/docker v20.10.0-beta1.0.20201110211921-af34b94a78a1+incompatible
	github.com/gabriel-vasile/mimetype v1.3.0
	github.com/klauspost/compress v1.8.3
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/mudler/go-pluggable v0.0.0-20210513155700-54c6443073af
	github.com/mudler/luet v0.0.0-20210601205410-5cccc34f32c0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.6.1
	github.com/vbatts/go-mtree v0.5.0
	golang.org/x/sys v0.0.0-20210603125802-9665404d3644 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/containerd/containerd v1.4.0-0 => github.com/containerd/containerd v1.4.0
	github.com/docker/docker v0.0.0 => github.com/docker/docker v20.10.5+incompatible
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
)
