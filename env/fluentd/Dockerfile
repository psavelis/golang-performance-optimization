FROM fluent/fluentd:v1.16-1
USER root
RUN gem install fluent-plugin-grafana-loki --no-document
COPY fluent.conf /fluentd/etc/fluent.conf
USER fluent