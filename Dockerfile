FROM scratch
ADD  it2s-socket-amd64 /
EXPOSE 4000
ENTRYPOINT ["./it2s-socket-amd64"]
