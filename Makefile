run:
	go build -o bin/tsuki-go -v .
	./bin/tsuki-go

git:
	git add .
	git commit -m "$(msg)"
	git push
