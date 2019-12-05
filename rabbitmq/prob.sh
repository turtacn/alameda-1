#!/usr/bin/env sh
#

if ! rabbitmqctl status > /dev/null 2>&1; then
    exit 1
fi

test_dedup_queues="model predict"
dedup_check_dir="/tmp/.dedup-check"
file="${dedup_check_dir}/prob.txt"

if [ ! -d "$dedup_check_dir" ]
then
    mkdir -p $dedup_check_dir
fi

if [ -f "$file" ]
then
    exit 0
else
    echo "$0: File '${file}' not found."
    dedup_worked="true"
    for test_queue in $test_dedup_queues; do
        count=`rabbitmqctl list_queues | grep "$test_queue\t" | awk '{print $2}'`
        if [ "$count" != "" ]; then
            continue
        fi
        MQ_USER=${RABBITMQ_DEFAULT_USER:-admin}
        MQ_PASSWD=${RABBITMQ_DEFAULT_PASS:-adminpass}
        ./usr/local/bin/rabbitmqcmd publish --queue=$test_queue
        ./usr/local/bin/rabbitmqcmd publish --queue=$test_queue
        count=`rabbitmqctl list_queues | grep "$test_queue\t" | awk '{print $2}'`
        if [ "$count" = "1" ]
        then
            rabbitmqctl purge_queue $test_queue
        else
            rabbitmqctl delete_queue $test_queue
            dedup_worked="false"
        fi
    done
    if [ $dedup_worked == "true" ]; then
        touch $file
        exit 0
    fi
fi
exit 1
