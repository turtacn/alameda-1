FROM docker.io/rabbitmq:3.8.1-management-alpine

COPY plugins/ plugins/
COPY init.sh /init.sh

RUN chmod 755 /init.sh && \
    rabbitmq-plugins enable rabbitmq_message_deduplication rabbitmq_tracing

CMD ["/init.sh"]

