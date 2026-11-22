run:
	mkdir -p bin && rm bin/* -f 
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/wgcli_linux_amd64_release . && upx -9 bin/wgcli_linux_amd64_release
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o bin/wgcli_windows_386_release.exe . && upx -9 bin/wgcli_windows_386_release.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/wgcli_windows_amd64_release.exe . && upx -9 bin/wgcli_windows_amd64_release.exe
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/wgcli_darwin_amd64_release . && upx -9 --force-macos bin/wgcli_darwin_amd64_release

a: # android
	go get golang.org/x/mobile/bind
	gomobile bind -target android -androidapi 24 -o android/app/libs/gomobile.aar  github.com/stevenzack/wgcli/gomobile