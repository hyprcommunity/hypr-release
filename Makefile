APP_NAME := hypr-release
GUI_APP := hypr-release-gui

# Go modül adı (go.mod içinde aynı olmalı)
MODULE := hypr-release

# Derleme çıktı klasörü
BIN_DIR := bin

# Go ayarları
GO := go
GO_FLAGS := -ldflags="-s -w"

# Varsayılan hedef
.PHONY: all
all: clean deps fmt vet build gui

# ----------------------
# Modül hazırlığı ve bağımlılıklar
.PHONY: deps
deps:
	@if [ ! -f go.mod ]; then \
		echo ">>> go.mod bulunamadı, oluşturuluyor..."; \
		$(GO) mod init $(MODULE); \
	fi
	$(GO) mod tidy

# ----------------------
# Kod biçimlendirme ve kalite
.PHONY: fmt vet test
fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./... -v

# ----------------------
# CLI derleme
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GO_FLAGS) -o $(BIN_DIR)/$(APP_NAME) ./api/...

# ----------------------
# GUI derleme (Fyne ile)
.PHONY: gui
gui:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GO_FLAGS) -o $(BIN_DIR)/$(GUI_APP) ./gui-api

# ----------------------
# Temizlik
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

# ----------------------
# Kurulum (isteğe bağlı)
.PHONY: install
install: all
	install -Dm755 $(BIN_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	install -Dm755 $(BIN_DIR)/$(GUI_APP) /usr/local/bin/$(GUI_APP)
	@echo "Kurulum tamamlandı: /usr/local/bin/$(APP_NAME), $(GUI_APP)"
