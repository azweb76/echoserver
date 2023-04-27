build:
	docker build  --platform linux/amd64 -t azweb76/echoserver .

up:
	docker-compose up --build

publish: build
	docker tag azweb76/echoserver azweb76/echoserver:latest
	docker push azweb76/echoserver
