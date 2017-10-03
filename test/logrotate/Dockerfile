FROM ubuntu:xenial

RUN apt update -qq
RUN apt install -y logrotate ruby curl inotify-tools

# # td-agent
RUN curl https://packages.treasuredata.com/GPG-KEY-td-agent | apt-key add - \
    && echo "deb http://packages.treasuredata.com/2/ubuntu/xenial/ xenial contrib" > /etc/apt/sources.list.d/treasure-data.list \
    && apt update -qq && apt install -y td-agent


ADD kinesis-agent-go.service /etc/systemd/system/

RUN mkdir -p /var/log/kinesis-agent-go
WORKDIR /var/log/kinesis-agent-go/

ADD config.yml ./
ADD fluent.conf ./
ADD logrotate.conf ./
ADD run.rb ./
ADD kinesis-agent-go ./
