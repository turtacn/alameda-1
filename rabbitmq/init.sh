#!/usr/bin/env sh
#

MQ_USER=${RABBITMQ_DEFAULT_USER:-admin}
MQ_PASSWD=${RABBITMQ_DEFAULT_PASS:-adminpass}

check_rabbitmq_status()
{
    out=`rabbitmqctl status`
    if [ "$?" == "0" ]; then
        return 0
    fi
    return 1
}

check_queue_dedup()
{
local queue_name=$1
echo "Checking queue $queue_name deduplication"
queue_size=`rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD list queues | grep " $queue_name " | cut -d '|' -f 3 | awk '{$1=$1;print}'`

if [ "$queue_size" == "" ]; then
    echo "no result"
	return
else
    get_queue_count=$((2 * $queue_size))
	rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD get queue=$queue_name count=$get_queue_count | grep $queue_name | cut -d '|' -f 5 | sort | awk '{$1=$1;print}' | sed -r 's/[ ]*"createTimestamp":[ ]*[0-9]*//' > /tmp/content
	lines=`cat /tmp/content | wc -l`
	uniq_content=`cat /tmp/content | sort | uniq > /tmp/uniq_content`
	uniq_lines=`cat /tmp/uniq_content | wc -l`
	if [ "$lines" != "$uniq_lines" ]; then
		echo "message duplicated. Try to recreate the queue"
		rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD delete queue name=$queue_name
		rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD declare queue name=$queue_name arguments='{"x-message-deduplication":true}'
		while read p; do
#		    echo "publish $p"
			message_md5=`echo '$p' | md5sum | cut -f1 -d" "`
			rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD publish exchange=amq.default routing_key=$queue_name payload='$p' properties='{"headers":{"x-deduplication-header":"$message_md5"}}'
		done </tmp/uniq_content
	fi
fi
}

do_crond()
{
    sleep_time=600
    while :; do
        sleep ${sleep_time}
        echo "doing crond jobs"
        #for qn in predict model; do
        #    check_queue_dedup $qn
        #done
        echo "done crond jobs. Sleeing ${sleep_time}"
    done
    exit 0
}

echo "Bring up rabbitmq-server"
rabbitmq-server &

while ! check_rabbitmq_status; do
    echo "Waiting rabbitmq server ready"
    sleep 10
done

if ! rabbitmqadmin declare user name=$MQ_USER password=$MQ_PASSWD tags=administrator; then
    echo "create username/password failed"
fi
retry=0
retryTime=30
while ! rabbitmqctl authenticate_user $MQ_USER $MQ_PASSWD > /dev/null 2>&1; do
    if [ $retry -ge $retryTime ];then
        exit 1
    fi
    if ! rabbitmqadmin declare user name=$MQ_USER password=$MQ_PASSWD tags=administrator; then
        echo "create username/password failed"
    fi
    retry=$((retry+1))
    sleep 5
done
rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD declare permission vhost=/ user=$MQ_USER configure='.*' write='.*' read='.*'
rabbitmqadmin -u $MQ_USER -p $MQ_PASSWD delete user name=guest
if [ "$TRACE_ENABLED" == "true" ]; then
    rabbitmqctl trace_on
    curl -i -u $MQ_USER:$MQ_PASSWD -H "content-type:application/json" -XPUT \
         http://localhost:15672/api/traces/%2f/trace \
         -d'{"format":"json","pattern":"#",
             "tracer_connection_username":"'$MQ_USER'", "tracer_connection_password":"'$MQ_PASSWD'"}'
fi

echo "Running daemon jobs"
do_crond
