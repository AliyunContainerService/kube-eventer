#!/usr/bin/env bash
echo "$TZ" >  /etc/timezone
cp /usr/share/zoneinfo/$TZ   /etc/localtime
/kube-eventer "$@"