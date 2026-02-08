tmpdir := `mktemp -d`
version := "0.0.1"
tardir := tmpdir / "untitled-dm-" + version
tarball := tardir + ".tar"
distdir := "dist" / "untitled-dm-" + version

build:
    CGO_ENABLED=0 go build -o bin ./... 

dist: build
    mkdir {{ tardir }}
    cp -a README.md LICENSE go.mod go.sum untitled-dm.go example/ {{ tardir }}
    tar cavf {{ tarball }}.gz --transform='s,{{ tmpdir }}/,,' --absolute-names {{ tardir }}
    mkdir -p {{ distdir }}
    cp {{ tarball }}.* bin/* {{ distdir }}
    rm -rf {{ tarball }}.* {{ tardir }}
    rmdir {{ tmpdir }}

sign-dist id: dist
    rm -f {{ distdir }}/sha256sums*
    cd {{ distdir }} && sha256sum * > sha256sums
    gpg -u '{{ id }}' --detach-sign {{ distdir }}/sha256sums
