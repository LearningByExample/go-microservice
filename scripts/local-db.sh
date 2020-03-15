#!/bin/bash

# Copyright (c) 2020 Learning by Example maintainers.
#
#  Permission is hereby granted, free of charge, to any person obtaining a copy
#  of this software and associated documentation files (the "Software"), to deal
#  in the Software without restriction, including without limitation the rights
#  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#  copies of the Software, and to permit persons to whom the Software is
#  furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included in
#  all copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
#  THE SOFTWARE.

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
