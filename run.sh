#!/bin/bash
############################################

# File Name : run.sh

# Purpose :

# Creation Date : 10-30-2017

# Last Modified : Mon 30 Oct 2017 02:44:02 AM UTC

# Created By : Kiyor 

############################################

groupadd -g 1001 nginx
useradd -u 1001 -g 1001 nginx
chown -R nginx:nginx .
while true; do
	echo "`date` playlist..."
	/root/gosu 1001:1001 /root/playlist
	echo "`date` sleep 300..."
	sleep 300
done
