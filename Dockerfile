FROM scratch

ARG TARGETPLATFORM=linux/amd64

COPY ${TARGETPLATFORM}/gdashlint /usr/local/bin/gdashlint

USER 65532:65532
WORKDIR /work

ENTRYPOINT ["/usr/local/bin/gdashlint"]
CMD ["--help"]
