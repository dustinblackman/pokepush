VERSION := 0.0.1
GLIDE_COMMIT := 91d42a717b7202c55568a7da05be915488253b8d
LINTER_COMMIT := 052c5941f855d3ffc9e8e8c446e0c0a8f0445410
.PHONY: build dev dev-build dist

build:
	go build -ldflags="-X main.version $(VERSION)" -o pokepush *.go

deps:
	@if [ "$$(which glide)" = "" ]; then \
		go get -v github.com/Masterminds/glide; \
		cd $$GOPATH/src/github.com/Masterminds/glide;\
		git checkout $(GLIDE_COMMIT);\
		go install;\
	fi
	glide install
	go install
	glide install

dev:
	which reflex && echo "" || go get github.com/cespare/reflex
	reflex -R '^vendor/' -r '\.go$\' -s -- sh -c 'go run *.go'

dev-build:
	which reflex && echo "" || go get github.com/cespare/reflex
	reflex -R '^vendor/' -r '\.go$\' -s -- sh -c 'go build -o pokepush *.go && ./pokepush'

docker:
	gox -os="linux" -arch="amd64" -output="pokepush" -ldflags="-X main.version $(VERSION)"
	chmod +x ./pokepush
	docker build -t dustinblackman/pokepush:latest .
	rm ./pokepush
	docker push dustinblackman/pokepush:latest

install: deps test
	go install -ldflags="-X main.version $(VERSION)" *.go

dist:
	which gox && echo "" || go get github.com/mitchellh/gox
	rm -rf tmp dist
	mkdir dist
	gox -output='tmp/{{.OS}}-{{.Arch}}-$(VERSION)/{{.Dir}}' -ldflags="-X main.version $(VERSION)"

	# Build for Windows
	@for i in $$(find ./tmp -type f -name "pokepush.exe" | awk -F'/' '{print $$3}'); \
	do \
	  zip -j "dist/pokepush-$$i.zip" "./tmp/$$i/pokepush.exe"; \
	done

	# Build for everything else
	@for i in $$(find ./tmp -type f -not -name "pokepush.exe" | awk -F'/' '{print $$3}'); \
	do \
	  chmod +x "./tmp/$$i/pokepush"; \
	  tar -zcvf "dist/pokepush-$$i.tar.gz" --directory="./tmp/$$i" "./pokepush"; \
	done

	rm -rf tmp

setup-linter:
	@if [ "$$(which gometalinter)" = "" ]; then \
		go get -v github.com/alecthomas/gometalinter; \
		cd $$GOPATH/src/github.com/alecthomas/gometalinter;\
		git checkout $(LINTER_COMMIT);\
		go install;\
		gometalinter --install;\
	fi

test:
	make setup-linter
	gometalinter --vendor --fast --dupl-threshold=100 --cyclo-over=25 --min-occurrences=5 --disable=gas ./...
