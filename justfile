@default:
    just --list

[group('admin')]
install:
    go install ./cmd/...

[group('example')]
testserver:
    go run ./cmd/ukuleleweb --store_dir=testdata/wiki --main_page=UkuleleWeb -md.shortlinks=man=https://man.gnoack.org/

[group('example')]
static:
    mkdir -p public
    go run ./cmd/uku static -out_dir=public -site_title="UkuleleWeb Demo" -url_style=flat testdata/wiki/*

[group('maintainer')]
update-golden:
    go test -run 'TestFullPageRendering|TestRenderStaticHTMLGolden' -update .

[group('maintainer')]
[working-directory: 'testdata/wiki']
@update-example-wiki:
    for f in *; do echo "${f}"; curl -s -o "${f}" "https://wiki.gnoack.org/${f}.md"; done
    find . -empty -type f -delete
