.PHONY: build ui dev test clean install-desktop uninstall-desktop

build: ui
	go build -o lockbox ./cmd/lockbox

ui:
	cd ui && npm install && npm run build

dev:
	cd ui && npm run dev

test:
	gofmt -l $$(find . -name '*.go' -not -path './ui/*')
	go vet ./...
	go test ./...
	cd ui && npx vue-tsc --noEmit

clean:
	rm -f lockbox
	rm -rf ui/dist internal/server/dist/index.html internal/server/dist/assets

install-desktop:
	./scripts/install-desktop.sh

uninstall-desktop:
	systemctl --user disable --now lockbox-ui.service 2>/dev/null || true
	rm -f ~/.config/systemd/user/lockbox-ui.service
	rm -f ~/.local/share/applications/lockbox.desktop
	rm -f ~/.local/share/icons/hicolor/scalable/apps/lockbox.svg
	rm -f ~/.local/bin/lockbox-web
	systemctl --user daemon-reload
	@echo "Desktop launcher removed (vault untouched)."
