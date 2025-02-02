# Makefile

# OSを検出
ifeq ($(OS),Windows_NT)
    BUILD_SCRIPT = scripts\build.bat
else
    BUILD_SCRIPT = scripts\build.sh
endif

# デフォルトのターゲット
.PHONY: build
build:
	@echo "Running build script: $(BUILD_SCRIPT)"
	$(BUILD_SCRIPT)
