default: testacc

ifdef BACKSTAGE_BASE_URL
BACKSTAGE_BASE_URL := $(BACKSTAGE_BASE_URL)
else
BACKSTAGE_BASE_URL := http://demo.backstage.io
endif

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 BACKSTAGE_BASE_URL=${BACKSTAGE_BASE_URL} go test ./... -v $(TESTARGS) -timeout 120m -cover
