VERSION=1.2.0

bins: dep
	cd make-static-appimage && go build
	cd ..
	cd static-appimage-runtime && go build

dep:
	go get -u ./...
	#go get

tools:
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr

ver:
	echo version $(VERSION)

gittag:
	git tag v$(VERSION)
	git push --tags origin master

clean:
	rm -rf dist

dist:
	mkdir -p dist

gox:
	cd make-static-appimage	 && CGO_ENABLED=0 gox -ldflags="-s -w" -output="../dist/{{.Dir}}_{{.OS}}_{{.Arch}}"
	cd static-appimage-runtime && CGO_ENABLED=0 gox -ldflags="-s -w" -output="../dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

draft:
	ghr -draft v$(VERSION) dist/



