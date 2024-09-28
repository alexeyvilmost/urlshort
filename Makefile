GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

BINDIR=/Users/study/study/urlshort/cmd/shortener

DBCONN="port=5432 user=app dbname=shortener password=app host=localhost"

FILENAME=/Users/study/study/urlshort/cmd/shortener/storage.csv

FLAGS=-f $(FILENAME) # -d $(DBCONN)

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.57.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint 


.PHONY: test-%
test-full-%:
	go build -C cmd/shortener -o shortener
	shortenertest \
	-test.v -test.run=^TestIteration$*$$ \
	-source-path=. \
	-binary-path=cmd/shortener/shortener \
    -file-storage-path=/Users/study/study/urlshort/cmd/shortener/storage.csv \
	-server-port=8080

test-%:
	shortenertest \
	-test.v -test.run=^TestIteration$*$$ \
	-source-path=. \
	-binary-path=cmd/shortener/shortener \
    -file-storage-path=/Users/study/study/urlshort/cmd/shortener/storage.csv \
	-server-port=8080

#	-database-dsn="port=5432 user=app dbname=shortener password=app sslmode=disable host=localhost"

.PHONY: build-n-run
build-n-run:
	go build -C cmd/shortener -o shortener
	./cmd/shortener/shortener $(FLAGS)