FROM busybox

EXPOSE 9876

WORKDIR /opt/orbs

VOLUME /opt/orbs/db

ADD ./_bin/ /opt/orbs/

CMD /opt/orbs/trash-panda
