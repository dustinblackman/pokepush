VERSION := 0.0.1
.PHONY: build dev dev-build dist

build:
	go build -o pokepush *.go

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

setup:
	which glide && echo "" || go get github.com/Masterminds/glide
	glide install

test:
	which gometalinter && echo "" || (go get github.com/alecthomas/gometalinter && gometalinter --install)
	gometalinter --vendor --fast --dupl-threshold=100 --cyclo-over=25 --min-occurrences=5 ./...
