build:
	go generate -v
	go install -v
	go build -v -ldflags "-X github.com/flownative/localbeach/pkg/version.Version=dev" -o beach

compile:
	go build -v -ldflags "-X github.com/flownative/localbeach/pkg/version.Version=dev" -o beach
