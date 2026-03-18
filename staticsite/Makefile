WIKI_DIR  := testdata/wiki
OUT_DIR   := public
TEMPLATE  := staticsite/page.html
PAGES     := $(wildcard $(WIKI_DIR)/*)
HTML      := $(patsubst $(WIKI_DIR)/%,$(OUT_DIR)/%.html,$(PAGES))
UKU       := uku

all: $(OUT_DIR)/static/style.css $(OUT_DIR)/static/favicon.svg $(HTML)

$(OUT_DIR)/%.html: $(WIKI_DIR)/% $(TEMPLATE)
	$(UKU) render -out.template=$(TEMPLATE) -wiki.base_url=/ -out.url_style=flat -wiki.title="UkuleleWeb Demo" $< > $@

$(OUT_DIR)/static/%: static/%
	@mkdir -p $(@D)
	cp $< $@

clean:
	rm -rf $(OUT_DIR)

.PHONY: all clean
