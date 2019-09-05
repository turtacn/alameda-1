FROM docker.io/rabbitmq:3.7.17-management-alpine

COPY plugins/ plugins/
COPY init.sh /init.sh

RUN chmod 755 /init.sh

RUN rabbitmq-plugins enable rabbitmq_message_deduplication

CMD ["/init.sh"]

