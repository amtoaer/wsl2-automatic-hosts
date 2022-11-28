.PHONY: clean version build package
VERSION=$(shell git describe --tags || echo "unknown")

clean:
	rm release/wah
	rm -rf release/src release/pkg

version:
	sed -i 's/_ver/${VERSION}/g' release/PKGBUILD

build:
	go build -ldflags="-s -w" -o release/wah ./main.go

package: version build
	cd release && makepkg && cd ..