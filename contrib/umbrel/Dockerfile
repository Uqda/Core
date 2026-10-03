# The wrapper adds only Umbrel control/UI glue to the verified release image.
FROM ghcr.io/uqda/core:26.0.5@sha256:96c13d2e2049c42ef88c28e07e62e2506ff0825260011750f1a38c2c143bbf58
RUN apk add --no-cache python3 && addgroup -g 1000 dashboard && adduser -D -u 1000 -G dashboard dashboard
COPY contrib/umbrel/control.py contrib/umbrel/dashboard.py contrib/umbrel/service_access.py /opt/uqda-umbrel/
COPY contrib/umbrel/web /opt/uqda-umbrel/web
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
WORKDIR /opt/uqda-umbrel
ENTRYPOINT ["python3"]
CMD ["control.py"]
