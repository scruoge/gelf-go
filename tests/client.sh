#!/usr/bin/env bash
#
# credits https://gist.github.com/gm3dmo/7721379
#
# This script can be used to raise a graylog2/gelf message
# gzip it and send it to a graylog server using netcat (nc)
grayloghost=127.0.0.1
hostname='gelftester'
short_message='test message short version'
full_message='longer test message. dont\n worry be happy'
level=1
facility='gelftester'
# gnu date
date=$(date +'%s.%N')
env_name='test'
app_name='testapp'

# Read the message into the variable ${gelf_message}
# see http://graylog2.org/gelf for mor info

read -r -d '' gelf_message <<EOF
{
  "version": "1.0",
  "host": "${hostname}",
  "short_message": "${short_message}",
  "full_message": "${full_message}",
  "timestamp": ${date},
  "level": ${level},
  "facility": "${facility}",
  "_user_id": 42,
  "_Environment": "${env_name}",
  "_AppName": "${app_name}"
}
EOF


echo Plain
echo -----
echo -n "${gelf_message}"
echo ==============
echo
echo


echo Gzipped and Gunzipped
echo ---------------------
echo -n "${gelf_message}"| gzip -c -f -| xxd -p
echo ==============
echo
echo


echo Send
echo ----
echo  "${gelf_message}"| gzip -c -f - | nc -w 1 -u $grayloghost 12201
echo ==============
echo
echo