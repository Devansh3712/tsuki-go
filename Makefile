run:
	go build .
	./tsuki-go

git:
	git add .
	git commit -m "$(msg)"
	git push
