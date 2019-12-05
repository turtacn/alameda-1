#!/usr/bin/env sh
#

queue_name='test_queue'

if ! rabbitmqctl status > /dev/null 2>&1; then
    exit 1
fi

file="/tmp/prob.txt"
if [ ! -f "$file" ]
then
    echo "$0: File '${file}' not found."
    MQ_USER=${RABBITMQ_DEFAULT_USER:-admin}
    MQ_PASSWD=${RABBITMQ_DEFAULT_PASS:-adminpass}
    ./usr/local/bin/rabbitmqcmd publish --queue=$queue_name
    ./usr/local/bin/rabbitmqcmd publish --queue=$queue_name
    rabbitmqctl list_queues
    count=`rabbitmqctl list_queues | grep $queue_name | awk '{print $2}'`
    if [ "$count" = "1" ]
    then
        rabbitmqctl delete_queue $queue_name
        touch $file
        exit 0
    else
        exit 1
    fi
fi
exit 0

