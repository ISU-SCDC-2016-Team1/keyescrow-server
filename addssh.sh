#!/bin/sh
PRIVATE_TOKEN=UYRcEQzJwmoaEiTHtRjf
url="http://gitlab/api/v3/user?private_token=$PRIVATE_TOKEN&sudo="
url="$url$1"
echo $url
data=`curl $url -H 'Host: gitlab'`
ID=`echo $data | python -c 'import sys; import json; print(json.load(sys.stdin)[sys.argv[1]])' id`
echo $ID

url="http://gitlab/api/v3/user/keys?private_token=$PRIVATE_TOKEN&sudo="
url="$url$1"
echo $url

curtime=`date +'%D'`

data=`curl -F "id=$!" -F "title=User Gen Key $curtime" -F key=$3" $url -H 'Host: gitlab'`
echo $data
