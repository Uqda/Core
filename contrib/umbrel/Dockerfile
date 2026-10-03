# The wrapper adds only Umbrel control/UI glue to the verified release image.
FROM ghcr.io/uqda/core:26.0.4@sha256:7a7567a4694b8c4951909da6ed3473d043e66e523e28eeef4e73718c5787b8f7
RUN apk add --no-cache python3 && addgroup -g 1000 dashboard && adduser -D -u 1000 -G dashboard dashboard
COPY contrib/umbrel/control.py contrib/umbrel/dashboard.py /opt/uqda-umbrel/
COPY contrib/umbrel/web /opt/uqda-umbrel/web
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
WORKDIR /opt/uqda-umbrel
ENTRYPOINT ["python3"]
CMD ["control.py"]
