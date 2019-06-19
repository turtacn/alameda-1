#!/bin/sh
#
[ "${AIHOME}" = "" ] && export AIHOME=/opt/alameda/apiserver

#
while :
do
    cd ${AIHOME}/bin
    ${AIHOME}/bin/apiserver run
    [ -f /tmp/.pause ] && sleep 300 || sleep 30
done

exit 0
