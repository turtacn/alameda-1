#!/bin/sh
#
[ "${AIHOME}" = "" ] && export AIHOME=/opt/alameda/datapipe

# create pwent entry because logrotate need the running uid inside /etc/passwd
MY_UID="`id -u`"
if [ "`cat /etc/passwd | awk -F: '{print $3}' | grep \"^${MY_UID}$\"`" = "" ]; then
    sed -e "s/alameda:/alameda2:/g" /etc/passwd > /tmp/.tmpfn; cat /tmp/.tmpfn > /etc/passwd; rm -f /tmp/.tmpfn
    echo "alameda:x:${MY_UID}:0:Federator.ai:${AIHOME}:/bin/sh" >> /etc/passwd
fi

cron_run_hourly()
{
    sleep_time=3600
    while :; do
        case "`date +%H`" in
            "00") # hour is 00
                logrotate -v -f /etc/logrotate.conf
                ;;
        esac
        sleep ${sleep_time}
    done
    exit 0
}

# start crond
cron_run_hourly &

# start main service
while :; do
    cd ${AIHOME}/bin
    ${AIHOME}/bin/datapipe run
    [ -f /tmp/.pause ] && sleep 300 || sleep 30
done

exit 0
