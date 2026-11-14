.PHONY: valid invalid

valid:
	go test ./... -tags validonly -race -count 1000

invalid:
	go test ./... -race -count 1000
