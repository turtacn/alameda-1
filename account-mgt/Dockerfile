FROM osixia/openldap:1.2.4

ENV AIHOME=/opt/alameda/openldap \
    LDAP_LOG_LEVEL="256" \
    LDAP_ORGANISATION="federatorai" \
    LDAP_DOMAIN="prophetstor.com" \
    LDAP_ADMIN_PASSWORD="password" \
    LDAP_CONFIG_PASSWORD="password"\
    LDAP_READONLY_USER="false" \
    LDAP_READONLY_USER_USERNAME="readonly" \
    LDAP_READONLY_USER_PASSWORD="readonly" \
    LDAP_RFC2307BIS_SCHEMA="false" \
    LDAP_BACKEND="mdb" \
    LDAP_TLS="false" \
    LDAP_TLS_CRT_FILENAME="ldap.crt" \
    LDAP_TLS_KEY_FILENAME="ldap.key" \
    LDAP_TLS_CA_CRT_FILENAME="ca.crt" \
    LDAP_TLS_ENFORCE="false" \
    LDAP_TLS_CIPHER_SUITE="SECURE256:+SECURE128:-VERS-TLS-ALL:+VERS-TLS1.2:-RSA:-DHE-DSS:-CAMELLIA-128-CBC:-CAMELLIA-256-CBC" \
    LDAP_TLS_VERIFY_CLIENT="demand" \
    LDAP_REPLICATION="false" \
    LDAP_REPLICATION_CONFIG_SYNCPROV="binddn=\"cn=admin,cn=config\" bindmethod=simple credentials=$LDAP_CONFIG_PASSWORD searchbase=\"cn=config\" type=refreshAndPersist retry=\"60 +\" timeout=1 starttls=critical" \
    LDAP_REPLICATION_DB_SYNCPROV="binddn=\"cn=admin,$LDAP_BASE_DN\" bindmethod=simple credentials=$LDAP_ADMIN_PASSWORD searchbase=\"$LDAP_BASE_DN\" type=refreshAndPersist interval=00:00:00:10 retry=\"60 +\" timeout=1 starttls=critical" \
    LDAP_REPLICATION_HOSTS="#PYTHON2BASH:['ldap://ldap-one-service', 'ldap://ldap-two-service']" \
    KEEP_EXISTING_CONFIG="false" \
    LDAP_REMOVE_CONFIG_AFTER_SETUP="true" \
    LDAP_SSL_HELPER_PREFIX="ldap"

WORKDIR ${AIHOME}
ADD ./ldif /container/service/slapd/assets/config/bootstrap/ldif/custom
ADD ./init.sh /init.sh
ADD ./xray.sh ${AIHOME}/bin/xray.sh
ADD ./container/environment/00-init /container/environment/00-init
ADD ./container/service/00-init /container/service/00-init

RUN set -x \
    && apt-get update && apt-get install -y --no-install-recommends vim \
    && apt-get autoclean && apt-get autoremove && rm -rf /var/lib/apt/lists/*

EXPOSE 389

# USER ${USER_UID}

## Origin run tool is defined as the image ENTRYPOINT (see Dockerfile). 
## It set environment and run startup scripts and images process. More information in the Advanced User Guide.
ENTRYPOINT ["/usr/bin/python", "-u", "/container/tool/run" ]
CMD [ "--copy-service" ]
