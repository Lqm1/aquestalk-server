# Makefile

# OSを検出
ifeq ($(OS),Windows_NT)
    BUILD_SCRIPT = scripts\build.bat
	RUN_SCRIPT = scripts\run.bat
else
    BUILD_SCRIPT = scripts/build.sh
	RUN_SCRIPT = scripts/run.sh
endif

# デフォルトのターゲット
.PHONY: build
build:
	@echo "Running build script: $(BUILD_SCRIPT)"
	$(BUILD_SCRIPT)

.PHONY: run
run:
	@echo "Running run script: $(RUN_SCRIPT)"
	$(RUN_SCRIPT)
