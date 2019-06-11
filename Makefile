test:
	ginkgo -r -nodes=4 -progress

start_postgres_docker:
	docker run --rm -p 5432:5432 --name postgres -e POSTGRES_PASSWORD= -d postgres:9.5

stop_postgres_docker:
	docker stop postgres
	docker rm postgres
