tmpdir := `mktemp -d`
version := "1.0.0"
tardir := tmpdir / "untitled-dm-" + version
tarball := tardir + ".tar"
distdir := "dist" / "untitled-dm-" + version
static_build := "yes"

build:
    CGO_ENABLED={{ if static_build == "yes" { "0" } else { "1" } }} go build -o bin ./... 

check-version:
    [[ $(./bin/untitled-dm -V | cut -d' ' -f2) = v{{ version }} ]]

dist: build check-version
    mkdir {{ tardir }}
    cp -a README.md LICENSE go.mod go.sum untitled-dm.go example/ {{ tardir }}
    tar cavf {{ tarball }}.gz --transform='s,{{ tmpdir }}/,,' --absolute-names {{ tardir }}
    mkdir -p {{ distdir }}
    cp {{ tarball }}.* bin/* {{ distdir }}
    rm -rf {{ tmpdir }}

sign-dist id: dist
    rm -f {{ distdir }}/sha256sums*
    cd {{ distdir }} && sha256sum * > sha256sums
    gpg -u '{{ id }}' --detach-sign {{ distdir }}/sha256sums
