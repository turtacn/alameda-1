#!/bin/sh
#
[ "${APPHOME}" = "" ] && export APPHOME=/opt/prophetstor/federatorai/grafana

## create pwent entry because logrotate need the running uid inside /etc/passwd
MY_UID="`id -u`"
if [ "`cat /etc/passwd | awk -F: '{print $3}' | grep \"^${MY_UID}$\"`" = "" ]; then
    sed -e "s/alameda:/alameda2:/g" /etc/passwd > /tmp/.tmpfn; cat /tmp/.tmpfn > /etc/passwd; rm -f /tmp/.tmpfn
    echo "alameda:x:${MY_UID}:0:Federator.ai:${APPHOME}:/bin/sh" >> /etc/passwd
fi

do_crond()
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

do_liveness()
{
    echo
    [ "$?" != "0" ] && echo "Failed in server liveness probe." && return $?
    return 0
}

do_readiness()
{
    echo
    [ "$?" != "0" ] && echo "Failed in server readiness probe." && return $?
    return 0
}

do_start()
{
    ## start crond
    # $0 crond &

    ## initialize
    rm -rf /var/lib/grafana/*
    cd /
    tar xzvf ${APPHOME}/alameda-dashboard.tgz
    cd -

    ## start main service
    sh -x /run.sh run &  # run original startup script
    ## update sqlite
    while :; do
        if [ "0`sqlite3 /var/lib/grafana/grafana.db \"select count(*) from user where login = 'admin';\"`" -ge 1 ]; then
            sqlite3 /var/lib/grafana/grafana.db "update user set help_flags1 = 1"
            [ "$?" = "0" ] && break
        fi
        echo "Waiting on setting values into /var/lib/grafana/grafana.db."
        sleep 2
    done
    
    # start nginx service
    nginx -g 'daemon off;' &

    ## update 
    while :; do
        [ -f /tmp/.pause ] && sleep 300 || sleep 30
    done
    return $?
}

show_usage()
{
    /bin/echo -e "\n\nUsage: $0 [crond|liveness|readiness|start]\n\n"
    exit 1
}

##
## Main
##

## start crond only
case "$1" in
    "crond")
        do_crond
        exit $?
        ;;
    "liveness")
        do_liveness
        exit $?
        ;;
    "readiness")
        do_readiness
        exit $?
        ;;
    "start")
        do_start
        ;;
    *)
        show_usage
        exit $?
        ;;
esac

exit 0
