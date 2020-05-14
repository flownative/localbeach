build:
	go generate -v
	go install -v
	go build -v -o beach
