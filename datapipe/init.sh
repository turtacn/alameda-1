#!/bin/sh
#
[ "${AIHOME}" = "" ] && export AIHOME=/opt/alameda/datapipe

#
while :
do
    cd ${AIHOME}/bin
    ${AIHOME}/bin/datapipe run
    [ -f /tmp/.pause ] && sleep 300 || sleep 30
done

cat /etc/passwd /etc/group
sleep 600
exit 0
