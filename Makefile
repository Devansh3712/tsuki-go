run:
	go build -o bin/tsuki -v .
	./bin/tsuki

git:
	git add .
	git commit -m "$(msg)"
	git push
