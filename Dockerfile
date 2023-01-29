FROM debian
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update
RUN apt-get -y install apt-utils
RUN apt-get install iptables -y
RUN apt-get install iproute2 -y
RUN apt-get install curl -y
RUN apt-get install gpg -y
RUN apt-get install ca-certificates -y
RUN update-ca-certificates
WORKDIR /app
COPY bin /app/
# ENTRYPOINT [ "/bin/bash" ]
ENTRYPOINT [ "/app/linux/amd64/AnalyticsTracker" ]