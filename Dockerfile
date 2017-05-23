FROM scratch
ADD  . /
EXPOSE 4000
ENTRYPOINT ["./main"]
