#!/bin/sh
#

MQ_USER=${RABBITMQ_DEFAULT_USER:-admin}
MQ_PASSWD=${RABBITMQ_DEFAULT_PASS:-adminpass}
MQ_PREDICT_QUEUE_NAME=${MQ_PREDICT_QUEUE_NAME:-predict}

check_rabbitmq_status()
{
    out=`rabbitmqctl status`
    return $?
}

check_queue_dedup()
{
queue_size=`rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD list queues | grep $MQ_PREDICT_QUEUE_NAME | cut -d '|' -f 3 | awk '{$1=$1;print}'`

if [ "$queue_size" == "" ]; then
    echo "no result"
	return
else
    get_queue_count=$((2 * $queue_size))
	rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD get queue=$MQ_PREDICT_QUEUE_NAME count=$get_queue_count | grep $MQ_PREDICT_QUEUE_NAME | cut -d '|' -f 5 | sort | awk '{$1=$1;print}' > /tmp/content
	lines=`cat /tmp/content | wc -l`
	uniq_content=`cat /tmp/content | sort | uniq > /tmp/uniq_content`
	uniq_lines=`cat /tmp/uniq_content | wc -l`
	if [ "$lines" != "$uniq_lines" ]; then
		echo "message duplicated. Try to recreate the queue"
		rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD delete queue name=$MQ_PREDICT_QUEUE_NAME
		rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD declare queue name=$MQ_PREDICT_QUEUE_NAME arguments='{"x-message-deduplication":true}'
		while read p; do
#		    echo "publish $p"
			message_md5=`echo '$p' | md5sum | cut -f1 -d" "`
			rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD publish exchange=amq.default routing_key=$MQ_PREDICT_QUEUE_NAME payload='$p' properties='{"headers":{"x-deduplication-header":"$message_md5"}}'
		done </tmp/uniq_content
	fi
fi
}

do_crond()
{
    sleep_time=3600
    while :; do
        sleep ${sleep_time}
        check_queue_dedup
    done
    exit 0
}

echo "Bring up rabbitmq-server"
rabbitmq-server &

while ! check_rabbitmq_status; do
    sleep 10
done
rabbitmqadmin declare user name=$MQ_USER password=$MQ_PASSWD tags=administrator
rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD declare permission vhost=/ user=$MQ_USER configure='.*' write='.*' read='.*'
rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD delete user name=guest

echo "Running daemon jobs"
do_crond

