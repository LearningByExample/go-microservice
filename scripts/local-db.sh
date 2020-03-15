#!/bin/bash

set -o errexit

print_usage() {
	echo " "
	echo "  $0 <option>"
	echo " "
	echo " where options are:"
	echo " "
	echo "  - $0 start"
	echo "  - $0 stop"
}

get_docker_ps() {
	DOCKERPS=""
	DOCKERPS=$(docker ps -a -q --filter="name=pet-postgres")
}

kill_postgres() {
	docker kill $DOCKERPS
	docker rm $DOCKERPS
}

if [ $# -ne 1 ]; then
	echo "Illegal number of parameters, usage : "
	print_usage
	exit 2
fi

if [ "$1" = "start" ]; then
	get_docker_ps
	if [[ ! -z "$DOCKERPS" ]]; then
		kill_postgres
	fi
	docker run --name pet-postgres -e POSTGRES_USER=petuser -e POSTGRES_PASSWORD=petpwd -e POSTGRES_DB=pets -d -p 5432:5432 postgres
else
	if [ "$1" = "stop" ]; then
		get_docker_ps
		if [[ -z "$DOCKERPS" ]]; then
			echo "postgres is not running"
		else
			kill_postgres
		fi
	else
		echo "Invalid option, usage : "
		print_usage
		exit 2
	fi
fi

exit 0
