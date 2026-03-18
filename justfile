set allow-duplicate-recipes

import? 'local.just'

@default:
    just --list

[group('admin')]
install:
    go install ./cmd/...

[group('example')]
testserver:
    go run ./cmd/ukuleleweb --store_dir=testdata/wiki --main_page=UkuleleWeb -md.shortlinks=man=https://man.gnoack.org/

[group('example')]
static wiki_dir="testdata/wiki" out_dir="public" template="staticsite/page.html":
    make -f staticsite/Makefile WIKI_DIR={{wiki_dir}} OUT_DIR={{out_dir}} TEMPLATE={{template}} UKU="go run ./cmd/uku"

[group('maintainer')]
update-golden:
    go test -run 'TestFullPageRendering|TestRenderStaticHTMLGolden' -update .

[group('maintainer')]
[working-directory: 'testdata/wiki']
@update-example-wiki:
    for f in *; do echo "${f}"; curl -s -o "${f}" "https://wiki.gnoack.org/${f}.md"; done
    find . -empty -type f -delete
