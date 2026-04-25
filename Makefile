DEMO_PREFIX := /tmp/himo-demo
DEMO_BIN    := $(DEMO_PREFIX)/bin/himo

.PHONY: tapes clean-tapes

tapes: tapes/quickstart.gif

tapes/quickstart.gif: tapes/quickstart.tape $(DEMO_BIN)
	PATH="$(DEMO_PREFIX)/bin:$$PATH" vhs $<

$(DEMO_BIN): $(shell find cmd internal -name '*.go')
	@mkdir -p $(dir $@)
	go build -o $@ ./cmd/himo

clean-tapes:
	rm -rf $(DEMO_PREFIX)/config $(DEMO_PREFIX)/state $(DEMO_PREFIX)/todos tapes/quickstart.gif
