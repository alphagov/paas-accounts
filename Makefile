test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r -nodes=4

start_postgres_docker:
	docker run --rm -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=postgres -d postgres:12

stop_postgres_docker:
	docker stop postgres
	docker rm postgres
