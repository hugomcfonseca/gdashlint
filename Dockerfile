FROM scratch

ARG TARGETPLATFORM

COPY ${TARGETPLATFORM}/gdashlint /usr/local/bin/gdashlint

USER 65532:65532
WORKDIR /work

ENTRYPOINT ["/usr/local/bin/gdashlint"]
CMD ["--help"]
